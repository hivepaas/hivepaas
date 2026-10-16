// Package licensekeys holds the public keys license keys are accepted from: the
// production set in prod/, and the development set in dev/, whose private halves
// are committed with tools/licensesign. A key id is its file name, as in
// releasekeys. See prod/README.md for how keys are added and rotated.
package licensekeys

import (
	"embed"
	"io/fs"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig/releasekeys"
)

// Each directory holds a README.md, so the patterns match even while it holds no
// key: go:embed refuses a pattern that matches nothing.
//
//go:embed prod/* dev/*
var embedded embed.FS

// Prod is the production keys, PEM by key id.
func Prod() (map[string][]byte, error) {
	return load(embedded, "prod")
}

// Dev is the development keys, PEM by key id.
func Dev() (map[string][]byte, error) {
	return load(embedded, "dev")
}

func load(fsys fs.FS, dir string) (map[string][]byte, error) {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	pems, err := releasekeys.Load(sub)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return pems, nil
}

// Trusted is the keys license keys are verified against. allowDev is whether a
// key signed by the development keys is accepted, which only development should
// say; the development keys are parsed either way, so such a key is refused for
// what it is.
//
// The production set is parsed only when it holds keys. Without them every
// production key is refused, which is the state of a binary built before the
// first pair was added.
func Trusted(allowDev bool) (*license.Trust, error) {
	return trusted(embedded, allowDev)
}

// trusted is Trusted with the keys read from fsys's prod/ and dev/.
func trusted(fsys fs.FS, allowDev bool) (*license.Trust, error) {
	prodPEMs, err := load(fsys, "prod")
	if err != nil {
		return nil, err
	}
	devPEMs, err := load(fsys, "dev")
	if err != nil {
		return nil, err
	}
	// A key id names one key, and the set a signature counts against is told by
	// it. An id in both sets is a build mistake, and a dangerous one: were it the
	// same key, whatever the committed development keys sign would pass as
	// production.
	for keyID := range prodPEMs {
		if _, found := devPEMs[keyID]; found {
			return nil, hperrors.Wrap(hperrors.ErrLicenseSigningKeyInvalid).
				WithExtraDetail("key id %q is both a production and a development key", keyID)
		}
	}

	trust := &license.Trust{AllowDev: allowDev}
	if len(prodPEMs) > 0 {
		if trust.Prod, err = license.Purpose.ParsePublicKeys(prodPEMs); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	if trust.Dev, err = license.Purpose.ParsePublicKeys(devPEMs); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return trust, nil
}
