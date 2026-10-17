package licensekeys

import (
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

// The embedded keys are otherwise exercised only by a key being pasted; a bad
// file would surface as every license refused in the field.
func TestEmbeddedKeysParse(t *testing.T) {
	dev, err := Dev()
	assert.NoError(t, err)
	assert.Contains(t, dev, "dev-ed")
	assert.Contains(t, dev, "dev-ml")
	_, err = license.Purpose.ParsePublicKeys(dev)
	assert.NoError(t, err)

	prod, err := Prod()
	assert.NoError(t, err)
	if len(prod) > 0 {
		_, err = license.Purpose.ParsePublicKeys(prod)
		assert.NoError(t, err)
	}
}

// productionKeys is the production set a release must carry, by key id: one key
// of each algorithm a license key is signed with. A rotation adds the new pair
// here as it adds its files to prod/, and removes the old one the same way.
var productionKeys = map[string]string{
	"lic-2026-ed": releasesig.AlgEd25519,
	"lic-2026-ml": releasesig.AlgMLDSA65,
}

// TestProductionKeysShipped is a release gate: the release workflow runs go test
// before it builds. A binary without the production pair refuses every key the
// portal issues, and runs every installation on Bee. A file missing, one more, or
// a key of the wrong algorithm under a pair's id stops the release.
func TestProductionKeysShipped(t *testing.T) {
	prod, err := Prod()
	if !assert.NoError(t, err) {
		return
	}
	keys, err := license.Purpose.ParsePublicKeys(prod)
	if !assert.NoError(t, err, "the production set must hold a key of every algorithm") {
		return
	}

	shipped := map[string]string{}
	for keyID, key := range keys {
		shipped[keyID] = key.Algorithm
	}
	assert.Equal(t, productionKeys, shipped)
}

// Whatever the committed development keys sign must never pass as production:
// not under a shared key id, and not as the same key under another id.
func TestNoProductionKeyIsADevelopmentKey(t *testing.T) {
	prod, err := Prod()
	assert.NoError(t, err)
	dev, err := Dev()
	assert.NoError(t, err)

	devKeys := map[string]string{}
	for keyID, content := range dev {
		block, _ := pem.Decode(content)
		if assert.NotNil(t, block, keyID) {
			devKeys[string(block.Bytes)] = keyID
		}
	}
	for keyID, content := range prod {
		assert.NotContains(t, dev, keyID, "key id %q is in both sets", keyID)
		block, _ := pem.Decode(content)
		if assert.NotNil(t, block, keyID) {
			assert.NotContains(t, devKeys, string(block.Bytes), "production key %q is a development key", keyID)
		}
	}
}

func TestTrusted(t *testing.T) {
	for _, allowDev := range []bool{true, false} {
		trust, err := Trusted(allowDev)
		if !assert.NoError(t, err) {
			continue
		}
		assert.Equal(t, allowDev, trust.AllowDev)
		assert.Len(t, trust.Dev, 2, "the development keys are parsed whether or not they count")

		prod, err := Prod()
		assert.NoError(t, err)
		assert.Len(t, trust.Prod, len(prod))
	}
}

func TestTrustedFrom(t *testing.T) {
	dev, err := Dev()
	assert.NoError(t, err)
	fsys := func(prod map[string][]byte) fstest.MapFS {
		files := fstest.MapFS{"prod/README.md": {Data: []byte("docs")}, "dev/README.md": {Data: []byte("docs")}}
		for keyID, content := range dev {
			files["dev/"+keyID+".pub.pem"] = &fstest.MapFile{Data: content}
		}
		for keyID, content := range prod {
			files["prod/"+keyID+".pub.pem"] = &fstest.MapFile{Data: content}
		}
		return files
	}

	trust, err := trusted(fsys(nil), false)
	assert.NoError(t, err)
	assert.Empty(t, trust.Prod, "no production keys yet: production keys are all refused")
	assert.Len(t, trust.Dev, 2)

	prodEd, prodML := newPublicPEMs(t)
	trust, err = trusted(fsys(map[string][]byte{"lic-2026-ed": prodEd, "lic-2026-ml": prodML}), true)
	assert.NoError(t, err)
	assert.Len(t, trust.Prod, 2)
	assert.Len(t, trust.Dev, 2)

	_, err = trusted(fsys(map[string][]byte{"dev-ed": prodEd, "lic-2026-ml": prodML}), false)
	assert.ErrorIs(t, err, hperrors.ErrLicenseSigningKeyInvalid, "a key id in both sets")

	_, err = trusted(fsys(map[string][]byte{"lic-2026-ed": prodEd}), false)
	assert.ErrorIs(t, err, hperrors.ErrLicenseSigningKeyInvalid, "a production set without an ML-DSA key")
}

// newPublicPEMs makes a fresh ed25519 and ML-DSA-65 public key, as PEM.
func newPublicPEMs(t *testing.T) (ed, ml []byte) {
	t.Helper()
	edPub, _, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	mlPriv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	assert.NoError(t, err)
	pemOf := func(pub any) []byte {
		der, err := x509.MarshalPKIXPublicKey(pub)
		assert.NoError(t, err)
		return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	}
	return pemOf(edPub), pemOf(mlPriv.PublicKey())
}
