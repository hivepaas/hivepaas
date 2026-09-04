// Command seedgen rewrites the encrypted values in the development seed.
//
// Seed data cannot be written by hand: every secret in it is ciphertext, and
// ciphertext depends on the key it was sealed with. This reads the seed, decrypts
// each value with the scheme it currently carries, and re-seals it with a data
// encryption key derived here - emitting the encryption_keys row to match.
//
// The dev app secret and data key are fixed constants on purpose: the seed has to
// be reproducible and reviewable in git, and it is development data.
//
// Usage:
//
//	go run ./tools/seedgen -in hivepaas_app/db/seed/seed.sql
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

const (
	// devAppSecret is what the seeded values used to be encrypted with, and what
	// the seeded data key is wrapped with. Development only.
	devAppSecret = "xxx"
	// devDataKeySeed produces the fixed data key the seed is sealed with, so
	// regenerating the seed does not churn every line.
	devDataKeySeed = "32-byte-long"
	// devEncryptionKeyID is the id of the seeded encryption_keys row.
	devEncryptionKeyID = "01JZZZZZZZZZZZZZZZZZZZZZZZ"
)

// saltPrefixed matches a value encrypted with the old per-value derived key.
var saltPrefixed = regexp.MustCompile(regexp.QuoteMeta(base.EncryptionSaltPrefix) + `[^"]+`)

func main() {
	inPath := flag.String("in", "", "path to seed.sql")
	flag.Parse()

	if *inPath == "" {
		fail("-in is required")
	}

	key, err := datakey.New([]byte(devDataKeySeed))
	if err != nil {
		fail("build the dev data key: %v", err)
	}

	content, err := os.ReadFile(*inPath)
	if err != nil {
		fail("read the seed: %v", err)
	}

	rewritten, converted, err := resealValues(string(content), key)
	if err != nil {
		fail("%v", err)
	}

	rewritten, err = ensureEncryptionKeyRow(rewritten, key)
	if err != nil {
		fail("%v", err)
	}

	if err := os.WriteFile(*inPath, []byte(rewritten), 0o600); err != nil {
		fail("write the seed: %v", err)
	}
	fmt.Printf("re-sealed %d values in %s\n", converted, *inPath)
}

// resealValues replaces every derived-key value with one sealed by the data key.
func resealValues(content string, key *datakey.Key) (string, int, error) {
	var failure error
	converted := 0

	rewritten := saltPrefixed.ReplaceAllStringFunc(content, func(match string) string {
		if failure != nil {
			return match
		}
		// HashField shares the prefix but holds a hash, not ciphertext. It fails to
		// decrypt, which is exactly how it is told apart - and left alone.
		plain, err := cryptoutil.DecryptBase64(match, devAppSecret)
		if err != nil {
			return match
		}

		sealed, err := key.Seal(plain)
		if err != nil {
			failure = fmt.Errorf("seal a value: %w", err)
			return match
		}
		converted++
		return sealed
	})
	return rewritten, converted, failure
}

// ensureEncryptionKeyRow prepends the row that holds the wrapped data key, which
// the app needs before it can read anything else in the seed.
func ensureEncryptionKeyRow(content string, key *datakey.Key) (string, error) {
	if strings.Contains(content, "INSERT INTO encryption_keys") {
		return content, nil
	}

	wrapped, err := key.Wrap(devAppSecret)
	if err != nil {
		return "", fmt.Errorf("wrap the dev data key: %w", err)
	}

	row := fmt.Sprintf(`-- The data encryption key every seeded secret is sealed with, wrapped by the
-- development app secret (%q). Regenerate with: go run ./tools/seedgen
INSERT INTO encryption_keys (id, wrapped_key, is_active, created_at, updated_at)
VALUES ('%s', '%s', TRUE, NOW(), NOW());

`, devAppSecret, devEncryptionKeyID, wrapped)

	return row + content, nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seedgen: "+format+"\n", args...)
	os.Exit(1)
}
