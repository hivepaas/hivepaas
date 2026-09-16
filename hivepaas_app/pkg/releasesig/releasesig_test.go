package releasesig

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type testKey struct {
	id   string
	alg  string
	sign func(data []byte, context string) []byte
	pub  crypto.PublicKey
}

func newEd25519(t *testing.T, id string) *testKey {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	return &testKey{id: id, alg: AlgEd25519, pub: pub, sign: func(data []byte, context string) []byte {
		sig, err := priv.Sign(nil, data, &ed25519.Options{Context: context})
		assert.NoError(t, err)
		return sig
	}}
}

func newMLDSA65(t *testing.T, id string) *testKey {
	t.Helper()
	priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	assert.NoError(t, err)
	return &testKey{id: id, alg: AlgMLDSA65, pub: priv.PublicKey(), sign: func(data []byte, context string) []byte {
		sig, err := priv.SignDeterministic(data, &mldsa.Options{Context: context})
		assert.NoError(t, err)
		return sig
	}}
}

func pemOf(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func trust(t *testing.T, keys ...*testKey) PublicKeys {
	t.Helper()
	pems := map[string][]byte{}
	for _, key := range keys {
		pems[key.id] = pemOf(t, key.pub)
	}
	parsed, err := ParsePublicKeys(pems)
	assert.NoError(t, err)
	return parsed
}

func entry(key *testKey, data []byte) Entry {
	return Entry{KeyID: key.id, Algorithm: key.alg, Sig: base64.StdEncoding.EncodeToString(key.sign(data, Context))}
}

func sigFile(t *testing.T, data []byte, entries ...Entry) []byte {
	t.Helper()
	sum := sha256.Sum256(data)
	content, err := json.Marshal(File{SHA256: hex.EncodeToString(sum[:]), Signatures: entries})
	assert.NoError(t, err)
	return content
}

func TestVerify(t *testing.T) {
	ed := newEd25519(t, "2026-ed")
	ml := newMLDSA65(t, "2026-ml")
	keys := trust(t, ed, ml)
	data := []byte(`{"stable":{"appVersion":"v0.1.1"}}`)
	changed := []byte(`{"stable":{"appVersion":"v0.1.0"}}`)

	t.Run("accepts both signatures by trusted keys", func(t *testing.T) {
		assert.NoError(t, Verify(keys, data, sigFile(t, data, entry(ed, data), entry(ml, data))))
	})

	t.Run("passes over a signature by a key it does not know", func(t *testing.T) {
		// A release signed with a key only newer binaries trust, during a rotation.
		next := newMLDSA65(t, "2027-ml")
		assert.NoError(t, Verify(keys, data,
			sigFile(t, data, entry(next, data), entry(ed, data), entry(ml, data))))
	})

	cases := map[string][]byte{
		"ed25519 signature missing": sigFile(t, data, entry(ml, data)),
		"ml-dsa signature missing":  sigFile(t, data, entry(ed, data)),
		"no signatures":             sigFile(t, data),
		"only unknown keys": sigFile(t, data,
			entry(newEd25519(t, "x-ed"), data), entry(newMLDSA65(t, "x-ml"), data)),
		"file changed after signing": sigFile(t, data, entry(ed, changed), entry(ml, changed)),
		"one of two signatures over another file": sigFile(t, data,
			entry(ed, data), entry(ml, changed)),
		"other key under a trusted id": sigFile(t, data,
			entry(newEd25519(t, ed.id), data), entry(ml, data)),
		"algorithm claimed differs from the key": sigFile(t, data,
			entry(ed, data), Entry{KeyID: ml.id, Algorithm: AlgEd25519, Sig: entry(ml, data).Sig}),
		"signed without context": sigFile(t, data,
			Entry{KeyID: ed.id, Algorithm: AlgEd25519, Sig: base64.StdEncoding.EncodeToString(ed.sign(data, ""))},
			entry(ml, data)),
		"signed for another purpose": sigFile(t, data,
			entry(ed, data),
			Entry{KeyID: ml.id, Algorithm: AlgMLDSA65,
				Sig: base64.StdEncoding.EncodeToString(ml.sign(data, "hivepaas-templates-v1"))}),
		"malformed signature": sigFile(t, data,
			Entry{KeyID: ed.id, Algorithm: AlgEd25519, Sig: "!!"}, entry(ml, data)),
		"sha256 field mismatch": func() []byte {
			content, _ := json.Marshal(File{
				SHA256:     hex.EncodeToString(make([]byte, sha256.Size)),
				Signatures: []Entry{entry(ed, data), entry(ml, data)},
			})
			return content
		}(),
		"too many signatures": func() []byte {
			entries := []Entry{entry(ed, data), entry(ml, data)}
			for range maxSignatures {
				entries = append(entries, Entry{KeyID: "filler", Algorithm: AlgEd25519})
			}
			return sigFile(t, data, entries...)
		}(),
		"malformed file": []byte(`not json`),
		"empty file":     nil,
	}
	for name, content := range cases {
		t.Run("refuses "+name, func(t *testing.T) {
			assert.ErrorIs(t, Verify(keys, data, content), hperrors.ErrReleaseSignatureInvalid)
		})
	}
}

func TestParsePublicKeys(t *testing.T) {
	ed := pemOf(t, newEd25519(t, "ed").pub)
	ml := pemOf(t, newMLDSA65(t, "ml").pub)

	keys, err := ParsePublicKeys(map[string][]byte{"ed": ed, "ml": ml})
	assert.NoError(t, err)
	assert.Equal(t, AlgEd25519, keys["ed"].Algorithm)
	assert.Equal(t, AlgMLDSA65, keys["ml"].Algorithm)

	ml44, err := mldsa.GenerateKey(mldsa.MLDSA44())
	assert.NoError(t, err)
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	for name, pems := range map[string]map[string][]byte{
		"no keys":             {},
		"no ml-dsa key":       {"ed": ed},
		"no ed25519 key":      {"ml": ml},
		"not PEM":             {"ed": ed, "ml": ml, "bad": []byte("<BASE64_PUBLIC_KEY>")},
		"ML-DSA-44":           {"ed": ed, "ml": ml, "ml44": pemOf(t, ml44.PublicKey())},
		"unsupported key":     {"ed": ed, "ml": ml, "ec": pemOf(t, ec.Public())},
		"private key in slot": {"ed": ed, "ml": pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1}})},
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			_, err := ParsePublicKeys(pems)
			assert.ErrorIs(t, err, hperrors.ErrReleaseSigningKeyInvalid)
		})
	}
}
