package hpappserviceimpl

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const releaseKeyFileSuffix = ".pub.pem"

// releaseKeysFS holds the public keys release.json is accepted from. See
// releasekeys/README.md for how keys are added and rotated.
//
//go:embed releasekeys
var releaseKeysFS embed.FS

// loadReleaseSigningKeys reads every <key-id>.pub.pem in fsys, by key id.
func loadReleaseSigningKeys(fsys fs.FS) (map[string][]byte, error) {
	matches, err := fs.Glob(fsys, "releasekeys/*"+releaseKeyFileSuffix)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	keys := make(map[string][]byte, len(matches))
	for _, path := range matches {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		keyID := strings.TrimSuffix(strings.TrimPrefix(path, "releasekeys/"), releaseKeyFileSuffix)
		keys[keyID] = content
	}
	return keys, nil
}
