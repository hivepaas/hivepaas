package settinginitserviceimpl

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
)

func certPEM(t *testing.T, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		NotBefore:    notAfter.AddDate(-1, 0, 0),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	assert.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestAdoptableSelfSignedUsesTheFilesOwnExpiry(t *testing.T) {
	timeNow := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	notAfter := timeNow.AddDate(0, 5, 0)

	gotNotAfter, ok := adoptableSelfSigned(certPEM(t, notAfter), timeNow)

	assert.True(t, ok)
	assert.Equal(t, notAfter, gotNotAfter)
}

func TestAdoptableSelfSignedRefusesWhatIsExpiredOrUnreadable(t *testing.T) {
	timeNow := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	_, ok := adoptableSelfSigned(certPEM(t, timeNow.Add(-time.Hour)), timeNow)
	assert.False(t, ok, "expired: made anew")

	_, ok = adoptableSelfSigned(certPEM(t, timeNow.Add(sslSelfSignedRenewBeforeExp/2)), timeNow)
	assert.False(t, ok, "inside the renewal window: made anew")

	_, ok = adoptableSelfSigned([]byte("not a certificate"), timeNow)
	assert.False(t, ok)
}
