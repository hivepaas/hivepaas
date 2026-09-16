package hpappserviceimpl

import (
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

type releaseSigner struct {
	edPriv ed25519.PrivateKey
	mlPriv *mldsa.PrivateKey
	keys   map[string][]byte
}

func newReleaseSigner(t *testing.T) *releaseSigner {
	t.Helper()
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	mlPriv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	assert.NoError(t, err)
	return &releaseSigner{
		edPriv: edPriv,
		mlPriv: mlPriv,
		keys: map[string][]byte{
			"test-ed": pemPublicKey(t, edPub),
			"test-ml": pemPublicKey(t, mlPriv.PublicKey()),
		},
	}
}

func (s *releaseSigner) sign(t *testing.T, data []byte) []byte {
	t.Helper()
	edSig, err := s.edPriv.Sign(nil, data, &ed25519.Options{Context: releasesig.Context})
	assert.NoError(t, err)
	mlSig, err := s.mlPriv.SignDeterministic(data, &mldsa.Options{Context: releasesig.Context})
	assert.NoError(t, err)
	sum := sha256.Sum256(data)
	content, err := json.Marshal(releasesig.Envelope{
		Payload: base64.StdEncoding.EncodeToString(data),
		SHA256:  hex.EncodeToString(sum[:]),
		Signatures: []releasesig.Entry{
			{KeyID: "test-ed", Algorithm: releasesig.AlgEd25519, Sig: base64.StdEncoding.EncodeToString(edSig)},
			{KeyID: "test-ml", Algorithm: releasesig.AlgMLDSA65, Sig: base64.StdEncoding.EncodeToString(mlSig)},
		},
	})
	assert.NoError(t, err)
	return content
}

func pemPublicKey(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func TestParseReleaseInfo(t *testing.T) {
	signer := newReleaseSigner(t)
	data := []byte(`{"stable":{"appVersion":"v999.0.0","appImage":"hivepaas/hivepaas:999.0.0"}}`)

	t.Run("signed release info is decoded", func(t *testing.T) {
		info, err := parseReleaseInfo(signer.sign(t, data), signer.keys)
		assert.NoError(t, err)
		assert.Equal(t, "hivepaas/hivepaas:999.0.0", info.Stable.AppImage)
		assert.True(t, info.Stable.CanUpdate)
	})

	t.Run("payload swapped after signing is refused", func(t *testing.T) {
		var env releasesig.Envelope
		assert.NoError(t, json.Unmarshal(signer.sign(t, data), &env))
		env.Payload = base64.StdEncoding.EncodeToString(
			[]byte(`{"stable":{"appVersion":"v999.0.0","appImage":"evil/hivepaas:999.0.0"}}`))
		tampered, err := json.Marshal(env)
		assert.NoError(t, err)

		info, err := parseReleaseInfo(tampered, signer.keys)
		assert.Nil(t, info)
		assert.ErrorIs(t, err, hperrors.ErrReleaseSignatureInvalid)
	})

	t.Run("unsigned release.json is refused", func(t *testing.T) {
		info, err := parseReleaseInfo(data, signer.keys)
		assert.Nil(t, info)
		assert.ErrorIs(t, err, hperrors.ErrReleaseSignatureInvalid)
	})

	t.Run("no trusted keys refuses everything", func(t *testing.T) {
		info, err := parseReleaseInfo(signer.sign(t, data), map[string][]byte{})
		assert.Nil(t, info)
		assert.ErrorIs(t, err, hperrors.ErrReleaseSigningKeyInvalid)
	})
}

func TestReleaseInfoURL(t *testing.T) {
	prev := config.Current()
	t.Cleanup(func() { config.SetCurrent(prev) })

	for env, want := range map[string]string{
		config.EnvDev:  "https://raw.githubusercontent.com/hivepaas/hivepaas/main/release.signed.json",
		config.EnvBeta: "https://raw.githubusercontent.com/hivepaas/hivepaas/release/release.signed.json",
		config.EnvProd: "https://raw.githubusercontent.com/hivepaas/hivepaas/release/release.signed.json",
	} {
		config.SetCurrent(&config.Config{Env: env})
		assert.Equal(t, want, releaseInfoURL(), env)
	}
}

func TestLoadReleaseSigningKeys(t *testing.T) {
	keys, err := loadReleaseSigningKeys(fstest.MapFS{
		"releasekeys/README.md":        {Data: []byte("docs")},
		"releasekeys/2026-ed.pub.pem":  {Data: []byte("ed")},
		"releasekeys/2026-ml.pub.pem":  {Data: []byte("ml")},
		"releasekeys/2026-ed.key":      {Data: []byte("never embedded as a key")},
		"releasekeys/nested/x.pub.pem": {Data: []byte("not at top level")},
		"elsewhere/2026-other.pub.pem": {Data: []byte("outside the directory")},
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string][]byte{"2026-ed": []byte("ed"), "2026-ml": []byte("ml")}, keys)
}

// The embedded keys are otherwise only ever exercised against the live
// release.json; a bad file would surface as "updates stopped working" in the field.
func TestEmbeddedReleaseSigningKeys(t *testing.T) {
	keys, err := loadReleaseSigningKeys(releaseKeysFS)
	assert.NoError(t, err)
	if len(keys) == 0 {
		t.Skip("no release signing keys in releasekeys/ yet")
	}
	_, err = releasesig.ParsePublicKeys(keys)
	assert.NoError(t, err)
}
