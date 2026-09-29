// Package releasekeys holds the public keys release.json is accepted from, for
// the app and for the installer's verifier (tools/releaseverify) alike: both
// read this one set, so they cannot trust different keys. See README.md for how
// keys are added and rotated.
package releasekeys

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const keyFileSuffix = ".pub.pem"

//go:embed *.pub.pem
var embedded embed.FS

// Embedded is the trusted keys, PEM by key id.
func Embedded() (map[string][]byte, error) {
	return Load(embedded)
}

// Load reads every <key-id>.pub.pem at the top of fsys, by key id.
func Load(fsys fs.FS) (map[string][]byte, error) {
	matches, err := fs.Glob(fsys, "*"+keyFileSuffix)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	keys := make(map[string][]byte, len(matches))
	for _, path := range matches {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		keys[strings.TrimSuffix(path, keyFileSuffix)] = content
	}
	return keys, nil
}

// Fingerprint names a set of keys, as sha256:<hex> over the key ids in order,
// each with its PEM. The installer records the fingerprint of the keys its
// verifier image carries, and a test holds it to this set, so the keys cannot
// be rotated here while installers go on verifying with the old ones.
func Fingerprint(pems map[string][]byte) string {
	ids := make([]string, 0, len(pems))
	for id := range pems {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	h := sha256.New()
	for _, id := range ids {
		h.Write([]byte(id + "\n"))
		h.Write(pems[id])
		h.Write([]byte("\n"))
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
