package x509util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/crypto/cryptobyte"
	cryptobyte_asn1 "golang.org/x/crypto/cryptobyte/asn1"
)

var (
	oidExtensionSubjectAltName = []int{2, 5, 29, 17}
	oidChallengePassword       = []int{1, 2, 840, 113549, 1, 9, 7}
)

type publicKeyInfo struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

type tbsCertificateRequest struct {
	Raw           asn1.RawContent
	Version       int
	Subject       asn1.RawValue
	PublicKey     publicKeyInfo
	RawAttributes []asn1.RawValue `asn1:"tag:0"`
}

type certificateRequest struct {
	Raw                asn1.RawContent
	TBSCSR             tbsCertificateRequest
	SignatureAlgorithm pkix.AlgorithmIdentifier
	SignatureValue     asn1.BitString
}

// CertificateRequest is the JSON representation of an X.509 certificate. It is
// used to build a certificate request from a template.
type CertificateRequest struct {
	Version            int                      `json:"version"`
	Subject            Subject                  `json:"subject"`
	DNSNames           MultiString              `json:"dnsNames"`
	EmailAddresses     MultiString              `json:"emailAddresses"`
	IPAddresses        MultiIP                  `json:"ipAddresses"`
	URIs               MultiURL                 `json:"uris"`
	SANs               []SubjectAlternativeName `json:"sans"`
	Extensions         []Extension              `json:"extensions"`
	SignatureAlgorithm SignatureAlgorithm       `json:"signatureAlgorithm"`
	ChallengePassword  string                   `json:"-"`
	PublicKey          interface{}              `json:"-"`
	PublicKeyAlgorithm x509.PublicKeyAlgorithm  `json:"-"`
	Signature          []byte                   `json:"-"`
	Signer             crypto.Signer            `json:"-"`
}

// NewCertificateRequest creates a certificate request from a template.
func NewCertificateRequest(signer crypto.Signer, opts ...Option) (*CertificateRequest, error) {
	pub := signer.Public()
	o, err := new(Options).apply(&x509.CertificateRequest{
		PublicKey: pub,
	}, opts)
	if err != nil {
		return nil, err
	}

	// If no template use only the certificate request with the default leaf key
	// usages.
	if o.CertBuffer == nil {
		return &CertificateRequest{
			PublicKey: pub,
			Signer:    signer,
		}, nil
	}

	// With templates
	var cr CertificateRequest
	if err := json.NewDecoder(o.CertBuffer).Decode(&cr); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling certificate")
	}
	cr.PublicKey = pub
	cr.Signer = signer

	// Generate the subjectAltName extension if the certificate contains SANs
	// that are not supported in the Go standard library.
	if cr.hasExtendedSANs() && !cr.hasExtension(oidExtensionSubjectAltName) {
		ext, err := createCertificateRequestSubjectAltNameExtension(cr, cr.Subject.IsEmpty())
		if err != nil {
			return nil, err
		}
		// Prepend extension to achieve a certificate as similar as possible to
		// the one generated by the Go standard library.
		cr.Extensions = append([]Extension{ext}, cr.Extensions...)
	}

	return &cr, nil
}

// NewCertificateRequestFromX509 creates a CertificateRequest from an
// x509.CertificateRequest.
//
// This method is used to create the template variable .Insecure.CR or to
// initialize the Certificate when no templates are used.
// NewCertificateRequestFromX509 will always ignore the SignatureAlgorithm
// because we cannot guarantee that the signer will be able to sign a
// certificate template if Certificate.SignatureAlgorithm is set.
func NewCertificateRequestFromX509(cr *x509.CertificateRequest) *CertificateRequest {
	// Set SubjectAltName extension as critical if Subject is empty.
	fixSubjectAltName(cr)
	return &CertificateRequest{
		Version:            cr.Version,
		Subject:            newSubject(cr.Subject),
		DNSNames:           cr.DNSNames,
		EmailAddresses:     cr.EmailAddresses,
		IPAddresses:        cr.IPAddresses,
		URIs:               cr.URIs,
		Extensions:         newExtensions(cr.Extensions),
		PublicKey:          cr.PublicKey,
		PublicKeyAlgorithm: cr.PublicKeyAlgorithm,
		Signature:          cr.Signature,
		// Do not enforce signature algorithm from the CSR, it might not
		// be compatible with the certificate signer.
		SignatureAlgorithm: 0,
	}
}

// GetCertificateRequest returns the signed x509.CertificateRequest.
func (c *CertificateRequest) GetCertificateRequest() (*x509.CertificateRequest, error) {
	cert := c.GetCertificate().GetCertificate()
	asn1Data, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject:            cert.Subject,
		DNSNames:           cert.DNSNames,
		IPAddresses:        cert.IPAddresses,
		EmailAddresses:     cert.EmailAddresses,
		URIs:               cert.URIs,
		ExtraExtensions:    cert.ExtraExtensions,
		SignatureAlgorithm: x509.SignatureAlgorithm(c.SignatureAlgorithm),
	}, c.Signer)
	if err != nil {
		return nil, errors.Wrap(err, "error creating certificate request")
	}

	// Prepend challenge password and sign again
	if c.ChallengePassword != "" {
		asn1Data, err = c.addChallengePassword(asn1Data)
		if err != nil {
			return nil, err
		}
	}

	// This should not fail
	return x509.ParseCertificateRequest(asn1Data)
}

// addChallengePassword unmarshals the asn1Data into a certificateRequest and
// creates a new one with the challengePassword.
func (c *CertificateRequest) addChallengePassword(asn1Data []byte) ([]byte, error) {
	// Marshal challengePassword to ans1.RawValue
	// Build challengePassword attribute (RFC 2985 section-5.4)
	var builder cryptobyte.Builder
	builder.AddASN1(cryptobyte_asn1.SEQUENCE, func(child *cryptobyte.Builder) {
		child.AddASN1ObjectIdentifier(oidChallengePassword)
		child.AddASN1(cryptobyte_asn1.SET, func(value *cryptobyte.Builder) {
			switch {
			case isPrintableString(c.ChallengePassword, true, true):
				value.AddASN1(cryptobyte_asn1.PrintableString, func(s *cryptobyte.Builder) {
					s.AddBytes([]byte(c.ChallengePassword))
				})
			case isUTF8String(c.ChallengePassword):
				value.AddASN1(cryptobyte_asn1.UTF8String, func(s *cryptobyte.Builder) {
					s.AddBytes([]byte(c.ChallengePassword))
				})
			default:
				value.SetError(errors.New("error marshaling challenge password: password is not valid"))
			}
		})
	})

	b, err := builder.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling challenge password")
	}
	challengePasswordAttr := asn1.RawValue{
		FullBytes: b,
	}

	// Parse certificate request
	var csr certificateRequest

	rest, err := asn1.Unmarshal(asn1Data, &csr)
	if err != nil {
		return nil, err
	} else if len(rest) != 0 {
		return nil, errors.New("error unmarshalling certificate request: trailing data")
	}

	sigAlgo := csr.SignatureAlgorithm
	tbsCSR := tbsCertificateRequest{
		Version:       csr.TBSCSR.Version,
		Subject:       csr.TBSCSR.Subject,
		PublicKey:     csr.TBSCSR.PublicKey,
		RawAttributes: csr.TBSCSR.RawAttributes,
	}

	// Prepend challengePassword attribute
	tbsCSR.RawAttributes = append([]asn1.RawValue{challengePasswordAttr}, tbsCSR.RawAttributes...)

	// Marshal tbsCertificateRequest
	tbsCSRContents, err := asn1.Marshal(tbsCSR)
	if err != nil {
		return nil, errors.Wrap(err, "error creating certificate request")
	}
	tbsCSR.Raw = tbsCSRContents

	// Get the hash used previously
	var hashFunc crypto.Hash
	found := false
	sigAlgoOID := sigAlgo.Algorithm
	for _, m := range signatureAlgorithmMapping {
		if sigAlgoOID.Equal(m.oid) {
			hashFunc = m.hash
			found = true
			break
		}
	}
	if !found {
		return nil, errors.Errorf("error creating certificate request: unsupported signature algorithm %s", sigAlgoOID)
	}

	// Sign tbsCertificateRequest
	signed := tbsCSRContents
	if hashFunc != 0 {
		h := hashFunc.New()
		h.Write(signed)
		signed = h.Sum(nil)
	}

	var signature []byte
	signature, err = c.Signer.Sign(rand.Reader, signed, hashFunc)
	if err != nil {
		return nil, errors.Wrap(err, "error creating certificate request")
	}

	// Build new certificate request and marshal
	asn1Data, err = asn1.Marshal(certificateRequest{
		TBSCSR:             tbsCSR,
		SignatureAlgorithm: sigAlgo,
		SignatureValue: asn1.BitString{
			Bytes:     signature,
			BitLength: len(signature) * 8,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating certificate request")
	}
	return asn1Data, nil
}

// GetCertificate returns the Certificate representation of the
// CertificateRequest.
//
// GetCertificate will not specify a SignatureAlgorithm, it's not possible to
// guarantee that the certificate signer can sign with the CertificateRequest
// SignatureAlgorithm.
func (c *CertificateRequest) GetCertificate() *Certificate {
	return &Certificate{
		Subject:            c.Subject,
		DNSNames:           c.DNSNames,
		EmailAddresses:     c.EmailAddresses,
		IPAddresses:        c.IPAddresses,
		URIs:               c.URIs,
		SANs:               c.SANs,
		Extensions:         c.Extensions,
		PublicKey:          c.PublicKey,
		PublicKeyAlgorithm: c.PublicKeyAlgorithm,
		SignatureAlgorithm: 0,
	}
}

// GetLeafCertificate returns the Certificate representation of the
// CertificateRequest, including KeyUsage and ExtKeyUsage extensions.
//
// GetLeafCertificate will not specify a SignatureAlgorithm, it's not possible
// to guarantee that the certificate signer can sign with the CertificateRequest
// SignatureAlgorithm.
func (c *CertificateRequest) GetLeafCertificate() *Certificate {
	keyUsage := x509.KeyUsageDigitalSignature
	if _, ok := c.PublicKey.(*rsa.PublicKey); ok {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	cert := c.GetCertificate()
	cert.KeyUsage = KeyUsage(keyUsage)
	cert.ExtKeyUsage = ExtKeyUsage([]x509.ExtKeyUsage{
		x509.ExtKeyUsageServerAuth,
		x509.ExtKeyUsageClientAuth,
	})
	return cert
}

// hasExtendedSANs returns true if the certificate contains any SAN types that
// are not supported by the golang x509 library (i.e. RegisteredID, OtherName,
// DirectoryName, X400Address, or EDIPartyName)
//
// See also https://datatracker.ietf.org/doc/html/rfc5280.html#section-4.2.1.6
func (c *CertificateRequest) hasExtendedSANs() bool {
	for _, san := range c.SANs {
		if !(san.Type == DNSType || san.Type == EmailType || san.Type == IPType || san.Type == URIType || san.Type == AutoType || san.Type == "") {
			return true
		}
	}
	return false
}

// hasExtension returns true if the given extension oid is in the certificate.
func (c *CertificateRequest) hasExtension(oid ObjectIdentifier) bool {
	for _, e := range c.Extensions {
		if e.ID.Equal(oid) {
			return true
		}
	}
	return false
}

// CreateCertificateRequest creates a simple X.509 certificate request with the
// given common name and sans.
func CreateCertificateRequest(commonName string, sans []string, signer crypto.Signer) (*x509.CertificateRequest, error) {
	dnsNames, ips, emails, uris := SplitSANs(sans)
	asn1Data, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:       dnsNames,
		IPAddresses:    ips,
		EmailAddresses: emails,
		URIs:           uris,
	}, signer)
	if err != nil {
		return nil, errors.Wrap(err, "error creating certificate request")
	}
	// This should not fail
	return x509.ParseCertificateRequest(asn1Data)
}

// fixSubjectAltName makes sure to mark the SAN extension to critical if the
// subject is empty.
func fixSubjectAltName(cr *x509.CertificateRequest) {
	if subjectIsEmpty(cr.Subject) {
		for i, ext := range cr.Extensions {
			if ext.Id.Equal(oidExtensionSubjectAltName) {
				cr.Extensions[i].Critical = true
			}
		}
	}
}
