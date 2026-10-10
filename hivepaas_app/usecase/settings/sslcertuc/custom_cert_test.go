package sslcertuc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// ownCert is a certificate of one's own for the names, and its key, as PEM.
func ownCert(t *testing.T, notAfter time.Time, names ...string) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: names[0]},
		DNSNames:     names,
		NotBefore:    notAfter.Add(-30 * 24 * time.Hour),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	assert.NoError(t, err)
	keyDER, err := x509.MarshalECPrivateKey(key)
	assert.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
}

func customCert(domain, certPEM, keyPEM string) *entity.SSLCert {
	return &entity.SSLCert{CertType: base.SSLCertTypeCustom, Domain: domain, Certificate: certPEM,
		PrivateKey: entity.NewEncryptedField(keyPEM)}
}

// A certificate brought by its owner is taken with its own expiry, read from
// it rather than typed in beside it.
func TestACustomCertificateExpiresWhenItSays(t *testing.T) {
	notAfter := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	certPEM, keyPEM := ownCert(t, notAfter, "shop.example.com")
	cert := customCert("shop.example.com", certPEM, keyPEM)

	assert.NoError(t, checkCustomCert(cert))
	assert.Equal(t, notAfter, cert.ExpireAt)
}

// One for every name under a domain covers each of them, and the wildcard
// domain itself.
func TestAWildcardCertificateCoversTheNamesUnderIt(t *testing.T) {
	certPEM, keyPEM := ownCert(t, time.Now().Add(time.Hour), "*.example.com")

	assert.NoError(t, checkCustomCert(customCert("shop.example.com", certPEM, keyPEM)))
	assert.NoError(t, checkCustomCert(customCert("*.example.com", certPEM, keyPEM)))
}

// What would leave a domain served with the proxy's default certificate is
// refused, and says why.
func TestACustomCertificateThatCannotServeItsDomainIsRefused(t *testing.T) {
	certPEM, keyPEM := ownCert(t, time.Now().Add(time.Hour), "shop.example.com")
	_, otherKey := ownCert(t, time.Now().Add(time.Hour), "shop.example.com")

	for name, tc := range map[string]struct {
		cert   *entity.SSLCert
		reason string
	}{
		"not a certificate": {customCert("shop.example.com", "hello", keyPEM), "not a PEM certificate"},
		"key of another":    {customCert("shop.example.com", certPEM, otherKey), "private key is not the certificate's"},
		"another domain":    {customCert("blog.example.com", certPEM, keyPEM), "does not cover blog.example.com"},
	} {
		err := checkCustomCert(tc.cert)
		assert.ErrorIs(t, err, hperrors.ErrSSLCertUnusable, name)
		assert.Contains(t, hperrors.Wrap(err).Build(translation.LangEn).Detail, tc.reason, name)
	}
}

// A certificate HivePaaS obtains is its own business.
func TestOnlyACustomCertificateIsChecked(t *testing.T) {
	cert := &entity.SSLCert{CertType: base.SSLCertTypeSelfSigned, Domain: "shop.example.com"}

	assert.NoError(t, checkCustomCert(cert))
}
