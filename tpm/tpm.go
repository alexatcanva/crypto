package tpm

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/google/go-attestation/attest"

	"go.step.sm/crypto/tpm/internal/open"
	"go.step.sm/crypto/tpm/simulator"
	"go.step.sm/crypto/tpm/storage"
)

// TPM models a Trusted Platform Module. It provides an abstraction
// over the google/go-tpm and google/go-attestation packages, allowing
// functionalities of these packages to be performed in a uniform manner.
// Besides that, it provides a transparent method for persisting TPM
// objects, so that referencing and using these is simplified.
type TPM struct {
	deviceName   string
	attestConfig *attest.OpenConfig
	attestTPM    *attest.TPM
	rwc          io.ReadWriteCloser
	lock         sync.RWMutex
	store        storage.TPMStore
	simulator    *simulator.Simulator
	downloader   *downloader
	info         *Info
	eks          []*EK
}

// NewTPMOption is used to provide options when instantiating a new
// instance of TPM.
type NewTPMOption func(t *TPM) error

// WithDeviceName is used to provide the `name` or path to the TPM
// device.
func WithDeviceName(name string) NewTPMOption {
	return func(t *TPM) error {
		t.deviceName = name
		return nil
	}
}

// WithStore is used to set the TPMStore implementation to use for
// persisting TPM objects, including AKs and Keys.
func WithStore(store storage.TPMStore) NewTPMOption {
	return func(t *TPM) error {
		if store == nil {
			store = storage.BlackHole() // prevent nil storage; no persistence
		}

		t.store = store
		return nil
	}
}

// WithDisableDownload disables EK certificates from being downloaded
// from online hosts.
func WithDisableDownload() NewTPMOption {
	return func(t *TPM) error {
		t.downloader.enabled = false
		return nil
	}
}

// New creates a new TPM instance. It takes `opts` to configure
// the instance.
func New(opts ...NewTPMOption) (*TPM, error) {
	tpm := &TPM{
		attestConfig: &attest.OpenConfig{TPMVersion: attest.TPMVersion20}, // default configuration for TPM attestation use cases
		store:        storage.BlackHole(),                                 // default storage doesn't persist anything // TODO(hs): make this in-memory storage instead?
		downloader:   &downloader{enabled: true, maxDownloads: 10},        // EK certificate download (if required) is enabled by default
	}

	for _, o := range opts {
		if err := o(tpm); err != nil {
			return nil, err
		}
	}

	return tpm, nil
}

// Open readies the TPM for usage and marks it as being
// in use. This makes using the instance safe for
// concurrent use.
func (t *TPM) Open(ctx context.Context) error {
	// prevent opening the TPM multiple times if Open is called
	// within the package multiple times.
	if isInternalCall(ctx) {
		return nil
	}

	// lock the TPM instance; it's in use now
	t.lock.Lock()

	if err := t.store.Load(); err != nil { // TODO(hs): load this once? Or abstract this away.
		return err
	}

	// if a simulator was set, use it as the backing TPM device.
	// The simulator is currently only used for testing.
	if t.simulator != nil {
		if t.attestTPM == nil {
			at, err := attest.OpenTPM(&attest.OpenConfig{
				TPMVersion:     attest.TPMVersion20,
				CommandChannel: t.simulator,
			})
			if err != nil {
				return fmt.Errorf("failed opening attest.TPM: %w", err)
			}
			t.attestTPM = at
		}
		t.rwc = t.simulator
	} else {
		// TODO(hs): when an internal call to Open is performed, but when
		// switching the "TPM implementation" to use between the two types,
		// there's a possibility of a nil pointer exception. At the moment,
		// the only "go-tpm" call is for GetRandom(), but this could change
		// in the future.
		if isGoTPMCall(ctx) {
			rwc, err := open.TPM(t.deviceName)
			if err != nil {
				return fmt.Errorf("failed opening TPM: %w", err)
			}
			t.rwc = rwc
		} else {
			// TODO(hs): attest.OpenTPM doesn't currently take into account the
			// device name provided. This doesn't seem to be an available option
			// to filter on currently?
			at, err := attest.OpenTPM(t.attestConfig)
			if err != nil {
				return fmt.Errorf("failed opening TPM: %w", err)
			}
			t.attestTPM = at
		}
	}

	return nil
}

// Close closes the TPM instance, cleaning up resources and
// marking it ready to be use again.
func (t *TPM) Close(ctx context.Context) {
	// prevent closing the TPM multiple times if Open is called
	// within the package multiple times.
	if isInternalCall(ctx) {
		return
	}

	// if simulation is enabled, closing the TPM simulator must not
	// happen, because re-opening it will result in a different instance,
	// resulting in issues when running multiple test operations in
	// sequence. Closing a simulator has to be done in the calling code,
	// meaning it has to happen at the end of the test.
	if t.simulator != nil {
		t.lock.Unlock()
		return
	}

	// clean up the attest.TPM
	if t.attestTPM != nil {
		err := t.attestTPM.Close()
		_ = err // TODO: handle error correctly (in defer)
		t.attestTPM = nil
	}

	// clean up the go-tpm rwc
	if t.rwc != nil {
		err := t.rwc.Close()
		_ = err // TODO: handle error correctly (in defer)
		t.rwc = nil
	}

	// mark the TPM as ready to be used again
	t.lock.Unlock()
}
