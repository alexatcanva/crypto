package tss2

import (
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// key generated using:
//
//	openssl genpkey -provider tpm2 -algorithm RSA -out rsa-key.pem
var rsaTSS2PEM = `-----BEGIN TSS2 PRIVATE KEY-----
MIICEgYGZ4EFCgEDoAMBAQECBEAAAAEEggEYARYAAQALAAYAcgAAABAAEAgAAAAA
AAEAxaLvqN+HkUsj83CQxxKwjB/OODnMrLzCly3917Fv0iSehV0sjkrW+9kpksYL
5nAdhT/EhfOGupmq8hpEghEFUvocarGf/CtT7ra/FE1d6JT8F6bPe9lIUNcUhObx
Msw8JU3f/uPNurlA/yXb8ZTlGSHoXUrYbiDF36EJUKm2nc/QkdB4SyXrAaUjTFAC
8ifqb3cvtx1KrLGhbP+dQ1A2ytK8FaMauhOGp3gIlAOOVwcRg/DvaxzVBlKRW75u
21r6/w5Wo71IdeEPB/qlCfY3CGvk3in6Wf27gjS+vVa/gx9N9sV7Sq/QX5rWnyol
x5cb6miCxLAWrchU80bGi0h+BwSB4ADeACCMbeh27b01dSmGrtftJnU9ZJaXr6tr
us/XbpEh+QY/NQAQZlRBv8rA2Gicwnf5gbZcBGVAOxp/EWZH+MVEKQOk4EVM+ACa
EQRWNdlr/A/9lPyOZIWw4tajkfKZgfr0ia0kAK4Cfb3cTnXnsAp59/kLG35JM2gQ
/Xvm/CAxtvmVNPF6OrmWhNhoOHbveEok2nVKyr0WM5WQSCnnZH4DpyHtA3bXa2rb
aPU0KW6T7K0wYsGI0FHC/3arTjlVmXC+mze86n32TkO0rhQiAyn8jS6tbUe6b3jW
x7YfKP6g
-----END TSS2 PRIVATE KEY-----`

// key generated using:
//
//	openssl ecparam -provider tpm2 -name prime256v1 -genkey -noout -out ec-key.pem
var p256TSS2PEM = `-----BEGIN TSS2 PRIVATE KEY-----
MIHwBgZngQUKAQOgAwEBAQIEQAAAAQRYAFYAIwALAAYAcgAAABAAEAADABAAIO7L
aur7h2XAyrZTA8g6QMksNNoUvkMZ4xnjVUSn3k+bACDHuNRZDInoD5Nts7WUos0k
Oe0/tF/HfhfSQTqsHo/rKgSBgAB+ACBr9xn6R2V13ErShb75o+EyMqtFsTysp24f
VNZ7IWfEBAAQebW0tKoqBmNr/ZGTt2jAEO/gdLEuZ+TkiWYf3h8Jcc0bUrj6lA9I
W6fVV4B/ZtnADx9/YGB9FBY8Bu07W7m+PorVTCbXFfOAFmSUg3eB0bgb2TRtFevZ
izcX
-----END TSS2 PRIVATE KEY-----`

// key generated using:
//
//	tpm2_createprimary -c primary.ctx
//	tpm2_create -C primary.ctx -G ecc -u obj.pub -r obj.priv
//	tpm2_encodeobject -C primary.ctx -u obj.pub -r obj.priv -o obj.pem
var p256EmptyAuthFalse = `-----BEGIN TSS2 PRIVATE KEY-----
MIHwBgZngQUKAQOgAwEBAAIEQAAAAQRYAFYAIwALAAYAcgAAABAAEAADABAAIIgq
1VllQRCT45GbLlp1Wud0jiSfojBwp1MYljWMw1T7ACAblgTFkwvSMnzpArA8GjVP
ULHy7pJubvS2W7TxmzclRQSBgAB+ACDVXoK8RpE5XxjBcfeHpip9Dz2j7AUj0oE1
RKlDg/+dYgAQd9c9mgioJc8wFL1zaU4viH1fq3fObbfZF/L8oLrLv6u3Pg8qeGzf
ePVypgEUeJGw68er7UZb4ZSVfoGId6KLX9JE7IwyBkRWLhBU3sLANdgjTqlXUhAD
mnYo
-----END TSS2 PRIVATE KEY-----`

func TestParsePrivateKey(t *testing.T) {
	parsePEM := func(s string) []byte {
		block, _ := pem.Decode([]byte(s))
		return block.Bytes
	}

	type args struct {
		derBytes []byte
	}
	tests := []struct {
		name      string
		args      args
		want      *TPMKey
		assertion assert.ErrorAssertionFunc
	}{
		{"ok rsa", args{parsePEM(rsaTSS2PEM)}, &TPMKey{
			Type:      oidLoadableKey,
			EmptyAuth: true,
			Parent:    1073741825,
			PublicKey: []byte{
				0x01, 0x16, 0x00, 0x01, 0x00, 0x0b, 0x00, 0x06, 0x00, 0x72, 0x00, 0x00, 0x00, 0x10, 0x00, 0x10,
				0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0xc5, 0xa2, 0xef, 0xa8, 0xdf, 0x87, 0x91, 0x4b,
				0x23, 0xf3, 0x70, 0x90, 0xc7, 0x12, 0xb0, 0x8c, 0x1f, 0xce, 0x38, 0x39, 0xcc, 0xac, 0xbc, 0xc2,
				0x97, 0x2d, 0xfd, 0xd7, 0xb1, 0x6f, 0xd2, 0x24, 0x9e, 0x85, 0x5d, 0x2c, 0x8e, 0x4a, 0xd6, 0xfb,
				0xd9, 0x29, 0x92, 0xc6, 0x0b, 0xe6, 0x70, 0x1d, 0x85, 0x3f, 0xc4, 0x85, 0xf3, 0x86, 0xba, 0x99,
				0xaa, 0xf2, 0x1a, 0x44, 0x82, 0x11, 0x05, 0x52, 0xfa, 0x1c, 0x6a, 0xb1, 0x9f, 0xfc, 0x2b, 0x53,
				0xee, 0xb6, 0xbf, 0x14, 0x4d, 0x5d, 0xe8, 0x94, 0xfc, 0x17, 0xa6, 0xcf, 0x7b, 0xd9, 0x48, 0x50,
				0xd7, 0x14, 0x84, 0xe6, 0xf1, 0x32, 0xcc, 0x3c, 0x25, 0x4d, 0xdf, 0xfe, 0xe3, 0xcd, 0xba, 0xb9,
				0x40, 0xff, 0x25, 0xdb, 0xf1, 0x94, 0xe5, 0x19, 0x21, 0xe8, 0x5d, 0x4a, 0xd8, 0x6e, 0x20, 0xc5,
				0xdf, 0xa1, 0x09, 0x50, 0xa9, 0xb6, 0x9d, 0xcf, 0xd0, 0x91, 0xd0, 0x78, 0x4b, 0x25, 0xeb, 0x01,
				0xa5, 0x23, 0x4c, 0x50, 0x02, 0xf2, 0x27, 0xea, 0x6f, 0x77, 0x2f, 0xb7, 0x1d, 0x4a, 0xac, 0xb1,
				0xa1, 0x6c, 0xff, 0x9d, 0x43, 0x50, 0x36, 0xca, 0xd2, 0xbc, 0x15, 0xa3, 0x1a, 0xba, 0x13, 0x86,
				0xa7, 0x78, 0x08, 0x94, 0x03, 0x8e, 0x57, 0x07, 0x11, 0x83, 0xf0, 0xef, 0x6b, 0x1c, 0xd5, 0x06,
				0x52, 0x91, 0x5b, 0xbe, 0x6e, 0xdb, 0x5a, 0xfa, 0xff, 0x0e, 0x56, 0xa3, 0xbd, 0x48, 0x75, 0xe1,
				0x0f, 0x07, 0xfa, 0xa5, 0x09, 0xf6, 0x37, 0x08, 0x6b, 0xe4, 0xde, 0x29, 0xfa, 0x59, 0xfd, 0xbb,
				0x82, 0x34, 0xbe, 0xbd, 0x56, 0xbf, 0x83, 0x1f, 0x4d, 0xf6, 0xc5, 0x7b, 0x4a, 0xaf, 0xd0, 0x5f,
				0x9a, 0xd6, 0x9f, 0x2a, 0x25, 0xc7, 0x97, 0x1b, 0xea, 0x68, 0x82, 0xc4, 0xb0, 0x16, 0xad, 0xc8,
				0x54, 0xf3, 0x46, 0xc6, 0x8b, 0x48, 0x7e, 0x07,
			},
			PrivateKey: []byte{
				0x00, 0xde, 0x00, 0x20, 0x8c, 0x6d, 0xe8, 0x76, 0xed, 0xbd, 0x35, 0x75, 0x29, 0x86, 0xae, 0xd7,
				0xed, 0x26, 0x75, 0x3d, 0x64, 0x96, 0x97, 0xaf, 0xab, 0x6b, 0xba, 0xcf, 0xd7, 0x6e, 0x91, 0x21,
				0xf9, 0x06, 0x3f, 0x35, 0x00, 0x10, 0x66, 0x54, 0x41, 0xbf, 0xca, 0xc0, 0xd8, 0x68, 0x9c, 0xc2,
				0x77, 0xf9, 0x81, 0xb6, 0x5c, 0x04, 0x65, 0x40, 0x3b, 0x1a, 0x7f, 0x11, 0x66, 0x47, 0xf8, 0xc5,
				0x44, 0x29, 0x03, 0xa4, 0xe0, 0x45, 0x4c, 0xf8, 0x00, 0x9a, 0x11, 0x04, 0x56, 0x35, 0xd9, 0x6b,
				0xfc, 0x0f, 0xfd, 0x94, 0xfc, 0x8e, 0x64, 0x85, 0xb0, 0xe2, 0xd6, 0xa3, 0x91, 0xf2, 0x99, 0x81,
				0xfa, 0xf4, 0x89, 0xad, 0x24, 0x00, 0xae, 0x02, 0x7d, 0xbd, 0xdc, 0x4e, 0x75, 0xe7, 0xb0, 0x0a,
				0x79, 0xf7, 0xf9, 0x0b, 0x1b, 0x7e, 0x49, 0x33, 0x68, 0x10, 0xfd, 0x7b, 0xe6, 0xfc, 0x20, 0x31,
				0xb6, 0xf9, 0x95, 0x34, 0xf1, 0x7a, 0x3a, 0xb9, 0x96, 0x84, 0xd8, 0x68, 0x38, 0x76, 0xef, 0x78,
				0x4a, 0x24, 0xda, 0x75, 0x4a, 0xca, 0xbd, 0x16, 0x33, 0x95, 0x90, 0x48, 0x29, 0xe7, 0x64, 0x7e,
				0x03, 0xa7, 0x21, 0xed, 0x03, 0x76, 0xd7, 0x6b, 0x6a, 0xdb, 0x68, 0xf5, 0x34, 0x29, 0x6e, 0x93,
				0xec, 0xad, 0x30, 0x62, 0xc1, 0x88, 0xd0, 0x51, 0xc2, 0xff, 0x76, 0xab, 0x4e, 0x39, 0x55, 0x99,
				0x70, 0xbe, 0x9b, 0x37, 0xbc, 0xea, 0x7d, 0xf6, 0x4e, 0x43, 0xb4, 0xae, 0x14, 0x22, 0x03, 0x29,
				0xfc, 0x8d, 0x2e, 0xad, 0x6d, 0x47, 0xba, 0x6f, 0x78, 0xd6, 0xc7, 0xb6, 0x1f, 0x28, 0xfe, 0xa0,
			},
		}, assert.NoError},
		{"ok ec", args{parsePEM(p256TSS2PEM)}, &TPMKey{
			Type:      oidLoadableKey,
			EmptyAuth: true,
			Parent:    1073741825,
			PublicKey: []byte{
				0x00, 0x56, 0x00, 0x23, 0x00, 0x0b, 0x00, 0x06, 0x00, 0x72, 0x00, 0x00, 0x00, 0x10, 0x00, 0x10,
				0x00, 0x03, 0x00, 0x10, 0x00, 0x20, 0xee, 0xcb, 0x6a, 0xea, 0xfb, 0x87, 0x65, 0xc0, 0xca, 0xb6,
				0x53, 0x03, 0xc8, 0x3a, 0x40, 0xc9, 0x2c, 0x34, 0xda, 0x14, 0xbe, 0x43, 0x19, 0xe3, 0x19, 0xe3,
				0x55, 0x44, 0xa7, 0xde, 0x4f, 0x9b, 0x00, 0x20, 0xc7, 0xb8, 0xd4, 0x59, 0x0c, 0x89, 0xe8, 0x0f,
				0x93, 0x6d, 0xb3, 0xb5, 0x94, 0xa2, 0xcd, 0x24, 0x39, 0xed, 0x3f, 0xb4, 0x5f, 0xc7, 0x7e, 0x17,
				0xd2, 0x41, 0x3a, 0xac, 0x1e, 0x8f, 0xeb, 0x2a,
			},
			PrivateKey: []byte{
				0x00, 0x7e, 0x00, 0x20, 0x6b, 0xf7, 0x19, 0xfa, 0x47, 0x65, 0x75, 0xdc, 0x4a, 0xd2, 0x85, 0xbe,
				0xf9, 0xa3, 0xe1, 0x32, 0x32, 0xab, 0x45, 0xb1, 0x3c, 0xac, 0xa7, 0x6e, 0x1f, 0x54, 0xd6, 0x7b,
				0x21, 0x67, 0xc4, 0x04, 0x00, 0x10, 0x79, 0xb5, 0xb4, 0xb4, 0xaa, 0x2a, 0x06, 0x63, 0x6b, 0xfd,
				0x91, 0x93, 0xb7, 0x68, 0xc0, 0x10, 0xef, 0xe0, 0x74, 0xb1, 0x2e, 0x67, 0xe4, 0xe4, 0x89, 0x66,
				0x1f, 0xde, 0x1f, 0x09, 0x71, 0xcd, 0x1b, 0x52, 0xb8, 0xfa, 0x94, 0x0f, 0x48, 0x5b, 0xa7, 0xd5,
				0x57, 0x80, 0x7f, 0x66, 0xd9, 0xc0, 0x0f, 0x1f, 0x7f, 0x60, 0x60, 0x7d, 0x14, 0x16, 0x3c, 0x06,
				0xed, 0x3b, 0x5b, 0xb9, 0xbe, 0x3e, 0x8a, 0xd5, 0x4c, 0x26, 0xd7, 0x15, 0xf3, 0x80, 0x16, 0x64,
				0x94, 0x83, 0x77, 0x81, 0xd1, 0xb8, 0x1b, 0xd9, 0x34, 0x6d, 0x15, 0xeb, 0xd9, 0x8b, 0x37, 0x17,
			},
		}, assert.NoError},
		{"ok emptyAuth false", args{parsePEM(p256EmptyAuthFalse)}, &TPMKey{
			Type:      oidLoadableKey,
			EmptyAuth: false,
			Parent:    1073741825,
			PublicKey: []byte{
				0x00, 0x56, 0x00, 0x23, 0x00, 0x0b, 0x00, 0x06, 0x00, 0x72, 0x00, 0x00, 0x00, 0x10, 0x00, 0x10,
				0x00, 0x03, 0x00, 0x10, 0x00, 0x20, 0x88, 0x2a, 0xd5, 0x59, 0x65, 0x41, 0x10, 0x93, 0xe3, 0x91,
				0x9b, 0x2e, 0x5a, 0x75, 0x5a, 0xe7, 0x74, 0x8e, 0x24, 0x9f, 0xa2, 0x30, 0x70, 0xa7, 0x53, 0x18,
				0x96, 0x35, 0x8c, 0xc3, 0x54, 0xfb, 0x00, 0x20, 0x1b, 0x96, 0x04, 0xc5, 0x93, 0x0b, 0xd2, 0x32,
				0x7c, 0xe9, 0x02, 0xb0, 0x3c, 0x1a, 0x35, 0x4f, 0x50, 0xb1, 0xf2, 0xee, 0x92, 0x6e, 0x6e, 0xf4,
				0xb6, 0x5b, 0xb4, 0xf1, 0x9b, 0x37, 0x25, 0x45,
			},
			PrivateKey: []byte{
				0x00, 0x7e, 0x00, 0x20, 0xd5, 0x5e, 0x82, 0xbc, 0x46, 0x91, 0x39, 0x5f, 0x18, 0xc1, 0x71, 0xf7,
				0x87, 0xa6, 0x2a, 0x7d, 0x0f, 0x3d, 0xa3, 0xec, 0x05, 0x23, 0xd2, 0x81, 0x35, 0x44, 0xa9, 0x43,
				0x83, 0xff, 0x9d, 0x62, 0x00, 0x10, 0x77, 0xd7, 0x3d, 0x9a, 0x08, 0xa8, 0x25, 0xcf, 0x30, 0x14,
				0xbd, 0x73, 0x69, 0x4e, 0x2f, 0x88, 0x7d, 0x5f, 0xab, 0x77, 0xce, 0x6d, 0xb7, 0xd9, 0x17, 0xf2,
				0xfc, 0xa0, 0xba, 0xcb, 0xbf, 0xab, 0xb7, 0x3e, 0x0f, 0x2a, 0x78, 0x6c, 0xdf, 0x78, 0xf5, 0x72,
				0xa6, 0x01, 0x14, 0x78, 0x91, 0xb0, 0xeb, 0xc7, 0xab, 0xed, 0x46, 0x5b, 0xe1, 0x94, 0x95, 0x7e,
				0x81, 0x88, 0x77, 0xa2, 0x8b, 0x5f, 0xd2, 0x44, 0xec, 0x8c, 0x32, 0x06, 0x44, 0x56, 0x2e, 0x10,
				0x54, 0xde, 0xc2, 0xc0, 0x35, 0xd8, 0x23, 0x4e, 0xa9, 0x57, 0x52, 0x10, 0x03, 0x9a, 0x76, 0x28,
			},
		}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrivateKey(tt.args.derBytes)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParsePrivateKey_marshal(t *testing.T) {
	modKey := func(fn func(key *TPMKey)) *TPMKey {
		fakeKey := &TPMKey{
			Type:       oidLoadableKey,
			EmptyAuth:  true,
			Parent:     1234,
			PublicKey:  []byte("pubkey"),
			PrivateKey: []byte("privkey"),
		}
		fn(fakeKey)
		return fakeKey
	}

	fakePolicy1 := TPMPolicy{CommandCode: 1, CommandPolicy: []byte("fake-policy-1")}
	fakePolicy2 := TPMPolicy{CommandCode: 2, CommandPolicy: []byte("fake-policy-2")}

	type args struct {
		key *TPMKey
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{"ok", args{&TPMKey{
			Type:       oidLoadableKey,
			EmptyAuth:  true,
			Parent:     1234,
			PublicKey:  []byte("pubkey"),
			PrivateKey: []byte("privkey"),
		}}, assert.NoError},
		{"ok importable key", args{modKey(func(key *TPMKey) {
			key.Type = oidImportableKey
		})}, assert.NoError},
		{"ok sealed key", args{modKey(func(key *TPMKey) {
			key.Type = oidSealedKey
		})}, assert.NoError},
		{"ok emptyAuth false", args{modKey(func(key *TPMKey) {
			key.EmptyAuth = false
		})}, assert.NoError},
		{"ok policy", args{modKey(func(key *TPMKey) {
			key.Policy = []TPMPolicy{fakePolicy1}
		})}, assert.NoError},
		{"ok policies", args{modKey(func(key *TPMKey) {
			key.Policy = []TPMPolicy{fakePolicy1, fakePolicy2}
		})}, assert.NoError},
		{"ok secret", args{modKey(func(key *TPMKey) {
			key.Secret = []byte("secret")
		})}, assert.NoError},
		{"ok authPolicy", args{modKey(func(key *TPMKey) {
			key.AuthPolicy = []TPMAuthPolicy{
				{Name: "auth", Policy: []TPMPolicy{fakePolicy1}},
			}
		})}, assert.NoError},
		{"ok authPolicies", args{modKey(func(key *TPMKey) {
			key.AuthPolicy = []TPMAuthPolicy{
				{Name: "auth-1", Policy: []TPMPolicy{fakePolicy1}},
				{Name: "auth-2", Policy: []TPMPolicy{fakePolicy1, fakePolicy2}},
			}
		})}, assert.NoError},
		{"ok all", args{&TPMKey{
			Type:      oidLoadableKey,
			EmptyAuth: true,
			Policy:    []TPMPolicy{fakePolicy1, fakePolicy2},
			Secret:    []byte("secret"),
			AuthPolicy: []TPMAuthPolicy{
				{Name: "auth-1", Policy: []TPMPolicy{fakePolicy1, fakePolicy2}},
				{Name: "auth-2", Policy: []TPMPolicy{fakePolicy1, fakePolicy2}},
			},
			Parent:     1234,
			PublicKey:  []byte("pubkey"),
			PrivateKey: []byte("privkey"),
		}}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			derBytes, err := MarshalPrivateKey(tt.args.key)
			if err != nil {
				t.Errorf("MarshalPrivateKey() error = %v", err)
				return
			}
			fmt.Printf("%x\n", derBytes)
			got, err := ParsePrivateKey(derBytes)
			tt.assertion(t, err)
			assert.Equal(t, tt.args.key, got)
		})
	}
}

func TestParsePrivateKey_noEmptyAuth(t *testing.T) {
	type tpmKeyTest struct {
		Type       asn1.ObjectIdentifier
		EmptyAuth  *bool           `asn1:"optional,explicit,tag:0"`
		Policy     []TPMPolicy     `asn1:"optional,explicit,tag:1"`
		Secret     []byte          `asn1:"optional,explicit,tag:2"`
		AuthPolicy []TPMAuthPolicy `asn1:"optional,explicit,tag:3"`
		Parent     int
		PublicKey  []byte
		PrivateKey []byte
	}

	type args struct {
		key tpmKeyTest
	}
	tests := []struct {
		name       string
		args       args
		wantPrefix []byte
		assertion  assert.ErrorAssertionFunc
	}{
		{"ok", args{tpmKeyTest{
			Type:       oidLoadableKey,
			Parent:     123,
			PublicKey:  []byte("public key"),
			PrivateKey: []byte("private key"),
		}}, []byte{0x30, 0x24, 0x6, 0x6, 0x67, 0x81, 0x5, 0xa, 0x1, 0x3, 0x2, 0x01, 0x7b}, assert.NoError},
		{"ok full", args{tpmKeyTest{
			Type: oidImportableKey,
			Policy: []TPMPolicy{
				{CommandCode: 1, CommandPolicy: []byte("fake-policy")},
			},
			Secret: []byte("secret"),
			AuthPolicy: []TPMAuthPolicy{{
				Name:   "auth",
				Policy: []TPMPolicy{{CommandCode: 1, CommandPolicy: []byte("fake-policy")}},
			}},
			Parent:     123,
			PublicKey:  []byte("public key"),
			PrivateKey: []byte("private key"),
		}}, []byte{0x30, 0x70, 0x6, 0x6, 0x67, 0x81, 0x5, 0xa, 0x1, 0x4, 0xa1, 0x18, 0x30}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.args.key

			derBytes, err := asn1.Marshal(key)
			require.NoError(t, err)

			prefix := derBytes[:len(tt.wantPrefix)]
			assert.Equal(t, tt.wantPrefix, prefix)

			got, err := ParsePrivateKey(derBytes)
			require.NoError(t, err)

			assert.Equal(t, &TPMKey{
				Type:       key.Type,
				EmptyAuth:  false,
				Policy:     key.Policy,
				Secret:     key.Secret,
				AuthPolicy: key.AuthPolicy,
				Parent:     key.Parent,
				PublicKey:  key.PublicKey,
				PrivateKey: key.PrivateKey,
			}, got)
		})
	}
}

func TestMarshalPrivateKey(t *testing.T) {
	type args struct {
		key *TPMKey
	}
	tests := []struct {
		name      string
		args      args
		want      []byte
		assertion assert.ErrorAssertionFunc
	}{
		{"ok", args{&TPMKey{
			Type:       oidLoadableKey,
			EmptyAuth:  true,
			Parent:     1234,
			PublicKey:  []byte("pubkey"),
			PrivateKey: []byte("privkey"),
		}}, []byte{
			0x30, 0x22,
			0x6, 0x6, 0x67, 0x81, 0x5, 0xa, 0x1, 0x3,
			0xa0, 0x3, 0x1, 0x1, 0xff,
			0x2, 0x2, 0x4, 0xd2,
			0x4, 0x6, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79,
			0x4, 0x7, 0x70, 0x72, 0x69, 0x76, 0x6b, 0x65, 0x79,
		}, assert.NoError},
		{"fail nil", args{nil}, nil, assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalPrivateKey(tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
