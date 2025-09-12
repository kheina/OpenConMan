package certs

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"
)

// GenerateCertificate creates a new certificate with the provided options.
// Returns the pem-encoded cert bytes and private key bytes
func GenerateCertificate(opt ...Option) ([]byte, []byte, error) {
	const op = "srv.GenerateCertificate"
	opts := getOpts(opt...)

	if len(opts.withName) == 0 {
		opts.withName = append(opts.withName, "conman")
	}

	pkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to generate ECDSA private key: %w", op, err)
	}

	template := &x509.Certificate{
		KeyUsage:              x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year ig
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			Country:    opts.withCountry,
			Province:   opts.withProvince,
			Locality:   opts.withLocality,
			CommonName: opts.withName[0],
		},
		DNSNames: opts.withName,
	}
	certder, err := x509.CreateCertificate(rand.Reader, template, template, &pkey.PublicKey, pkey)

	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to generate certificate: %w", op, err)
	}
	var certpem bytes.Buffer
	pem.Encode(&certpem, &pem.Block{Type: "CERTIFICATE", Bytes: certder})

	pkeyder, err := x509.MarshalECPrivateKey(pkey)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to marshal pkey: %w", op, err)
	}
	var pkeypem bytes.Buffer
	pem.Encode(&pkeypem, &pem.Block{Type: "EC PRIVATE KEY", Bytes: pkeyder})
	return certpem.Bytes(), pkeypem.Bytes(), nil

	// cert, err := tls.X509KeyPair(certpem.Bytes(), pkeypem.Bytes())
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("%s: failed to load certificate: %w", op, err)
	// }
	// return &cert, nil, nil
}
