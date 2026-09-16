package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"slices"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

func newKeys(t *testing.T, prefix string) []*privateKey {
	t.Helper()
	_, ed, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ml, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		t.Fatal(err)
	}
	return []*privateKey{
		{id: prefix + "-ed", alg: algEd25519, key: ed},
		{id: prefix + "-ml", alg: algMLDSA65, key: ml},
	}
}

func publicOf(keys []*privateKey) []*publicKey {
	pubs := make([]*publicKey, 0, len(keys))
	for _, key := range keys {
		pubs = append(pubs, &publicKey{id: key.id, alg: key.alg, key: key.key.Public()})
	}
	return pubs
}

func TestSignOpen(t *testing.T) {
	keys := newKeys(t, "2026")
	data := []byte(`{"stable":{"appVersion":"v0.1.1"}}`)

	env, err := Sign(keys, data)
	if err != nil {
		t.Fatal(err)
	}
	content := FormatEnvelope(env)
	opened, err := Open(publicOf(keys), content)
	if err != nil {
		t.Fatalf("a fresh envelope must open: %v", err)
	}
	if string(opened) != string(data) {
		t.Fatal("the envelope must carry the signed file byte for byte")
	}

	changed := *env
	changed.Payload = base64.StdEncoding.EncodeToString([]byte(`{"stable":{"appVersion":"v0.1.0"}}`))
	if _, err = Open(publicOf(keys), FormatEnvelope(&changed)); err == nil {
		t.Fatal("a changed payload must not open")
	}
	if _, err = Open(publicOf(newKeys(t, "2026")), content); err == nil {
		t.Fatal("other keys under the same ids must not open it")
	}
	if _, err = Open(publicOf(keys[:1]), content); err == nil {
		t.Fatal("trusting only one algorithm's key must not satisfy both")
	}

	if _, err = Sign(keys[:1], data); err == nil {
		t.Fatal("signing with one algorithm only must be refused")
	}
	if _, err = Sign(append(keys, newKeys(t, "x")[0]), data); err == nil {
		t.Fatal("two keys of one algorithm must be refused")
	}
}

// The tool cannot import the app's verifier (it is kept stdlib-only), so the two
// copies of the format are held together here: what the tool signs, the app
// accepts, through the PEM files a real release would use.
func TestSignatureAcceptedByApp(t *testing.T) {
	if signContext != releasesig.Context {
		t.Fatalf("signContext %q differs from releasesig.Context %q", signContext, releasesig.Context)
	}
	if algEd25519 != releasesig.AlgEd25519 || algMLDSA65 != releasesig.AlgMLDSA65 {
		t.Fatal("algorithm names differ from releasesig")
	}
	if !slices.Equal(requiredAlgorithms, releasesig.RequiredAlgorithms()) {
		t.Fatalf("requiredAlgorithms %v differ from releasesig %v", requiredAlgorithms, releasesig.RequiredAlgorithms())
	}

	keys := newKeys(t, "2026")
	data := []byte(`{"stable":{"appVersion":"v0.1.1"}}`)
	env, err := Sign(keys, data)
	if err != nil {
		t.Fatal(err)
	}

	pems := map[string][]byte{}
	for _, key := range keys {
		pems[key.id] = pemPublicKey(t, key.key.Public())
	}
	trusted, err := releasesig.ParsePublicKeys(pems)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := releasesig.Open(trusted, FormatEnvelope(env))
	if err != nil {
		t.Fatalf("the app must accept what the tool signs: %v", err)
	}
	if string(opened) != string(data) {
		t.Fatal("the app must get back the file the tool signed")
	}
}

func TestFormatEnvelope_OneFieldPerLine(t *testing.T) {
	env, err := Sign(newKeys(t, "2026"), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(FormatEnvelope(env)), "\n")
	onOneLine := func(parts ...string) bool {
		return slices.ContainsFunc(lines, func(line string) bool {
			for _, part := range parts {
				if !strings.Contains(line, part) {
					return false
				}
			}
			return true
		})
	}
	// scripts/release-sign.sh reads these with sed.
	if !onOneLine(`"payload": "`+env.Payload+`"`) {
		t.Fatal("payload is not on one line")
	}
	for _, entry := range env.Signatures {
		if !onOneLine(`{"keyId":"`+entry.KeyID+`","alg":"`+entry.Algorithm+`","sig":"`+entry.Sig+`"}`) {
			t.Fatalf("signature by %q is not on one line", entry.KeyID)
		}
	}
}

func pemPublicKey(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}
