package sslserviceimpl

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (s *service) GenerateCertAsPEM(
	subject *pkix.Name,
	keyType base.SSLKeyType,
	notBefore, notAfter time.Time,
	isCA bool,
) (cert, key []byte, err error) {
	var priv any
	switch keyType {
	case base.SSLKeyTypeECP256:
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case base.SSLKeyTypeECP384:
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case base.SSLKeyTypeECP521:
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case base.SSLKeyTypeRSA2048:
		priv, err = rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
	case base.SSLKeyTypeRSA3072:
		priv, err = rsa.GenerateKey(rand.Reader, 3072) //nolint:mnd
	case base.SSLKeyTypeRSA4096:
		priv, err = rsa.GenerateKey(rand.Reader, 4096) //nolint:mnd
	default:
		return nil, nil, fmt.Errorf("%w: unrecognized key type '%v'",
			apperrors.ErrUnsupported, keyType)
	}
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) //nolint:mnd
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               *subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	var publicKey any
	switch k := priv.(type) {
	case *ecdsa.PrivateKey:
		publicKey = &k.PublicKey
	case *rsa.PrivateKey:
		publicKey = &k.PublicKey
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	outCertBuf := bytes.NewBuffer(make([]byte, 0, 1000)) //nolint:mnd
	if err := pem.Encode(outCertBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode certificate: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	outPrivBuf := bytes.NewBuffer(make([]byte, 0, 1000)) //nolint:mnd
	if err := pem.Encode(outPrivBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	return outCertBuf.Bytes(), outPrivBuf.Bytes(), nil
}
