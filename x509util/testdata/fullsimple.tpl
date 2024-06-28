{
    "version": 3,
    "subject": "subjectCommonName",
    "issuer": "issuerCommonName",
    "serialNumber": "0x1234567890",
    "dnsNames": "doe.com",
    "emailAddresses": "jane@doe.com",
    "ipAddresses": "127.0.0.1",
    "uris": "https://doe.com",
    "sans": [{"type":"dns", "value":"www.doe.com"}],
    "notBefore": "2009-02-13T23:31:30Z",
    "notAfter": "2009-02-14T23:31:30Z",
    "extensions": [{"id":"1.2.3.4","critical":true,"value":"ZXh0ZW5zaW9u"}],
    "keyUsage": ["digitalSignature"],
    "extKeyUsage": ["serverAuth"],
    "unknownExtKeyUsage": ["1.3.6.1.4.1.44924.1.6", "1.3.6.1.4.1.44924.1.7"],
    "subjectKeyId": "c3ViamVjdEtleUlk",
    "authorityKeyId": "YXV0aG9yaXR5S2V5SWQ=",
    "ocspServer": "https://ocsp.server",
    "issuingCertificateURL": "https://ca.com",
    "crlDistributionPoints": "https://ca.com/ca.crl",
    "policyIdentifiers": "1.2.3.4.5.6",
    "basicConstraints": {
        "isCA": false, 
        "maxPathLen": 0
    },
    "nameConstraints": {
        "critical": true,
        "permittedDNSDomains": "jane.doe.com",
        "excludedDNSDomains": "john.doe.com",
        "permittedIPRanges": "127.0.0.1/32",
        "excludedIPRanges": "0.0.0.0/0",
        "permittedEmailAddresses": "jane@doe.com",
        "excludedEmailAddresses": "john@doe.com",
        "permittedURIDomains": "https://jane.doe.com",
        "excludedURIDomains": "https://john.doe.com"
    },
    "signatureAlgorithm": "Ed25519"
}