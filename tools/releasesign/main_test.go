package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/x509"
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

func TestSignVerify(t *testing.T) {
	keys := newKeys(t, "2026")
	data := []byte(`{"stable":{"appVersion":"v0.1.1"}}`)

	sigFile, err := Sign(keys, data)
	if err != nil {
		t.Fatal(err)
	}
	if err = Verify(publicOf(keys), data, sigFile); err != nil {
		t.Fatalf("a fresh signature file must verify: %v", err)
	}

	if Verify(publicOf(keys), []byte(`{"stable":{"appVersion":"v0.1.0"}}`), sigFile) == nil {
		t.Fatal("a changed file must not verify")
	}
	if Verify(publicOf(newKeys(t, "2026")), data, sigFile) == nil {
		t.Fatal("other keys under the same ids must not verify")
	}
	if Verify(publicOf(keys[:1]), data, sigFile) == nil {
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
	sigFile, err := Sign(keys, data)
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
	if err = releasesig.Verify(trusted, data, FormatSigFile(sigFile)); err != nil {
		t.Fatalf("the app must accept what the tool signs: %v", err)
	}
}

func TestFormatSigFile_OneSignaturePerLine(t *testing.T) {
	sigFile, err := Sign(newKeys(t, "2026"), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(FormatSigFile(sigFile)), "\n")
	for _, entry := range sigFile.Signatures {
		found := slices.ContainsFunc(lines, func(line string) bool {
			return strings.Contains(line, `"keyId":"`+entry.KeyID+`"`) && strings.Contains(line, entry.Sig)
		})
		if !found {
			t.Fatalf("signature by %q is not on one line; scripts/release-sign.sh relies on that", entry.KeyID)
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
