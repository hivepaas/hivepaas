package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

var sealedValue = regexp.MustCompile(regexp.QuoteMeta(base.EncryptionKeyPrefix) + `[^"']+`)

// seedPath is the seed this tool maintains, relative to the tool directory.
const seedPath = "../../hivepaas_app/db/seed/seed.sql"

// wrappedKeyInSeed pulls the wrapped data key out of the encryption_keys row.
var wrappedKeyInSeed = regexp.MustCompile(`INSERT INTO encryption_keys[^;]+VALUES \('[^']+', '([^']+)'`)

// The seed has to be usable by an app started with the development app secret:
// unwrap the key it carries, then open every value sealed with it. A seed that
// fails this bricks every local database it is loaded into.
func TestSeedIsReadableWithTheDevAppSecret(t *testing.T) {
	content, err := os.ReadFile(filepath.Clean(seedPath))
	if err != nil {
		t.Fatalf("read the seed: %v", err)
	}

	match := wrappedKeyInSeed.FindStringSubmatch(string(content))
	if match == nil {
		t.Fatal("the seed has no encryption_keys row; run: go run ./tools/seedgen")
	}

	key, err := datakey.Unwrap(match[1], devAppSecret)
	if err != nil {
		t.Fatalf("the dev app secret does not unwrap the seeded key: %v", err)
	}

	sealed := sealedValue.FindAllString(string(content), -1)
	if len(sealed) == 0 {
		t.Fatal("the seed carries no sealed values, which cannot be right")
	}
	for _, value := range sealed {
		if _, err := key.Open(value); err != nil {
			t.Errorf("a seeded value does not open with the seeded key: %v", err)
		}
	}
	t.Logf("opened %d sealed values", len(sealed))
}

// Regenerating must not touch the API key hash: it shares the older prefix but is
// a hash, not ciphertext, and re-sealing it would break every seeded API key.
func TestSeedKeepsHashedValues(t *testing.T) {
	content, err := os.ReadFile(filepath.Clean(seedPath))
	if err != nil {
		t.Fatalf("read the seed: %v", err)
	}

	if !strings.Contains(string(content), `"secretKey": "`+base.EncryptionSaltPrefix) {
		t.Error("the seeded API key hash is gone; seedgen must leave hashed values alone")
	}
}

// Running the generator again must be a no-op, so regenerating does not churn
// every line in the diff.
func TestResealValuesIsIdempotent(t *testing.T) {
	content, err := os.ReadFile(filepath.Clean(seedPath))
	if err != nil {
		t.Fatalf("read the seed: %v", err)
	}

	key, err := datakey.New([]byte(devDataKeySeed))
	if err != nil {
		t.Fatalf("build the dev key: %v", err)
	}

	rewritten, converted, err := resealValues(string(content), key)
	if err != nil {
		t.Fatalf("reseal: %v", err)
	}
	if converted != 0 {
		t.Errorf("re-sealed %d values on an already generated seed", converted)
	}
	if rewritten != string(content) {
		t.Error("regenerating an already generated seed changed it")
	}
}
