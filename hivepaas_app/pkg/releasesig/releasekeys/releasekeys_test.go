package releasekeys

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

func TestLoadReadsTheTopLevelPublicKeys(t *testing.T) {
	keys, err := Load(fstest.MapFS{
		"README.md":        {Data: []byte("docs")},
		"2026-ed.pub.pem":  {Data: []byte("ed")},
		"2026-ml.pub.pem":  {Data: []byte("ml")},
		"2026-ed.key":      {Data: []byte("never embedded as a key")},
		"nested/x.pub.pem": {Data: []byte("not at top level")},
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string][]byte{"2026-ed": []byte("ed"), "2026-ml": []byte("ml")}, keys)
}

// The embedded keys are otherwise only ever exercised against the live
// release.json; a bad file would surface as "updates stopped working" in the field.
func TestEmbeddedKeysParse(t *testing.T) {
	keys, err := Embedded()
	assert.NoError(t, err)
	_, err = releasesig.ParsePublicKeys(keys)
	assert.NoError(t, err)
}

// The fingerprint names a set of keys: the same set gives the same one whatever
// the order it is read in, and any change to a key or a key id gives another.
func TestFingerprintNamesTheSetOfKeys(t *testing.T) {
	a := map[string][]byte{"2026-ed": []byte("ed"), "2026-ml": []byte("ml")}
	same := map[string][]byte{"2026-ml": []byte("ml"), "2026-ed": []byte("ed")}

	assert.Equal(t, Fingerprint(a), Fingerprint(same))
	assert.Regexp(t, `^sha256:[0-9a-f]{64}$`, Fingerprint(a))
	assert.NotEqual(t, Fingerprint(a), Fingerprint(map[string][]byte{"2026-ed": []byte("ed"), "2026-ml": []byte("m2")}))
	assert.NotEqual(t, Fingerprint(a), Fingerprint(map[string][]byte{"2027-ed": []byte("ed"), "2026-ml": []byte("ml")}))
	assert.NotEqual(t, Fingerprint(a), Fingerprint(map[string][]byte{"2026-ed": []byte("ed")}))
}
