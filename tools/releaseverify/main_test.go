package main

import (
	"bytes"
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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig/releasekeys"
)

// must stops the test on err. testify's require is not vendored in the main
// checkout, so this stands in for it.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type signer struct {
	id, alg string
	pub     crypto.PublicKey
	sign    func([]byte) []byte
}

func signers(t *testing.T) (map[string][]byte, []signer) {
	t.Helper()
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	must(t, err)
	ml, err := mldsa.GenerateKey(mldsa.MLDSA65())
	must(t, err)
	keys := []signer{
		{id: "t-ed", alg: releasesig.AlgEd25519, pub: edPub, sign: func(data []byte) []byte {
			sig, err := edPriv.Sign(nil, data, &ed25519.Options{Context: releasesig.Context})
			must(t, err)
			return sig
		}},
		{id: "t-ml", alg: releasesig.AlgMLDSA65, pub: ml.PublicKey(), sign: func(data []byte) []byte {
			sig, err := ml.SignDeterministic(data, &mldsa.Options{Context: releasesig.Context})
			must(t, err)
			return sig
		}},
	}
	pems := map[string][]byte{}
	for _, k := range keys {
		der, err := x509.MarshalPKIXPublicKey(k.pub)
		must(t, err)
		pems[k.id] = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	}
	return pems, keys
}

func envelopeOf(t *testing.T, data []byte, keys ...signer) []byte {
	t.Helper()
	sum := sha256.Sum256(data)
	env := releasesig.Envelope{Payload: base64.StdEncoding.EncodeToString(data), SHA256: hex.EncodeToString(sum[:])}
	for _, k := range keys {
		env.Signatures = append(env.Signatures, releasesig.Entry{
			KeyID: k.id, Algorithm: k.alg, Sig: base64.StdEncoding.EncodeToString(k.sign(data)),
		})
	}
	content, err := json.Marshal(env)
	must(t, err)
	return content
}

func runWith(t *testing.T, pems map[string][]byte, stdin []byte, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, bytes.NewReader(stdin), &stdout, &stderr, pems)
	return code, stdout.String(), stderr.String()
}

func TestAVerifiedReleaseIsWrittenOut(t *testing.T) {
	pems, keys := signers(t)
	release := []byte(`{"beta":{"appVersion":"v1.0.0-beta1"}}`)

	code, out, _ := runWith(t, pems, envelopeOf(t, release, keys...))

	assert.Equal(t, 0, code)
	assert.Equal(t, string(release), out)
}

func TestAReleaseWithoutBothSignaturesIsRefused(t *testing.T) {
	pems, keys := signers(t)

	code, out, errOut := runWith(t, pems, envelopeOf(t, []byte(`{"beta":{}}`), keys[0]))

	assert.Equal(t, 1, code)
	assert.Empty(t, out, "nothing of an unverified release reaches the installer")
	assert.NotEmpty(t, errOut)
}

func TestAReleaseSignedByOtherKeysIsRefused(t *testing.T) {
	pems, _ := signers(t)
	_, others := signers(t)

	code, out, _ := runWith(t, pems, envelopeOf(t, []byte(`{"beta":{}}`), others...))

	assert.Equal(t, 1, code)
	assert.Empty(t, out)
}

func TestAnAlteredPayloadIsRefused(t *testing.T) {
	pems, keys := signers(t)
	env := envelopeOf(t, []byte(`{"beta":{"appImage":"a"}}`), keys...)
	var parsed releasesig.Envelope
	must(t, json.Unmarshal(env, &parsed))
	forged := []byte(`{"beta":{"appImage":"b"}}`)
	sum := sha256.Sum256(forged)
	parsed.Payload, parsed.SHA256 = base64.StdEncoding.EncodeToString(forged), hex.EncodeToString(sum[:])
	altered, err := json.Marshal(parsed)
	must(t, err)

	code, out, _ := runWith(t, pems, altered)

	assert.Equal(t, 1, code)
	assert.Empty(t, out)
}

func TestTheFingerprintIsOfTheKeysItTrusts(t *testing.T) {
	pems, _ := signers(t)

	code, out, _ := runWith(t, pems, nil, "-fingerprint")

	assert.Equal(t, 0, code)
	assert.Equal(t, releasekeys.Fingerprint(pems)+"\n", out)
}

// The installer names the fingerprint of the keys its verifier image carries. A
// rotation here that did not rebuild the image and update install.sh would leave
// installers verifying with keys the app no longer trusts: this is where it fails.
func TestTheInstallerNamesTheFingerprintOfTheTrustedKeys(t *testing.T) {
	script, err := os.ReadFile("../../deployment/release/install.sh")
	must(t, err)
	pems, err := releasekeys.Embedded()
	must(t, err)

	want := "VERIFY_KEYS=" + releasekeys.Fingerprint(pems)
	assert.True(t, strings.Contains(string(script), "\n"+want+"\n"),
		"install.sh must have the line %s: rebuild the verifier image and update VERIFY_IMAGE and VERIFY_KEYS", want)
}
