package entity

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestRemapRefsRewritesObjectIDFields(t *testing.T) {
	data := &SSLCert{
		Domain:       "example.com",
		Provider:     ObjectID{ID: "old-provider"},
		AcmeProvider: ObjectID{ID: "old-acme"},
	}
	err := RemapRefs(data, map[string]string{
		"old-provider": "new-provider",
		"old-acme":     "new-acme",
	})
	assert.NoError(t, err)
	assert.Equal(t, "new-provider", data.Provider.ID)
	assert.Equal(t, "new-acme", data.AcmeProvider.ID)
	assert.Equal(t, "example.com", data.Domain, "non-reference fields must not move")
}

// SystemBackupCloudStorage is a bespoke struct holding a bare `ID string`, not
// an ObjectID. A walk keyed on the ObjectID type misses it; this is the
// regression guard for exactly that.
func TestRemapRefsRewritesBareIDFieldsInBespokeStructs(t *testing.T) {
	data := &SystemBackup{
		CloudStorage: SystemBackupCloudStorage{
			ID:     "old-storage",
			Bucket: "backups",
		},
	}
	err := RemapRefs(data, map[string]string{"old-storage": "new-storage"})
	assert.NoError(t, err)
	assert.Equal(t, "new-storage", data.CloudStorage.ID)
	assert.Equal(t, "backups", data.CloudStorage.Bucket)
}

func TestRemapRefsRewritesSlicesAndNestedPointers(t *testing.T) {
	data := &AppRoutingSettings{
		Port: 8080,
		Domains: []*AppDomain{
			{Domain: "a.example.com", SSLCert: ObjectID{ID: "cert-1"}},
			{Domain: "b.example.com", SSLCert: ObjectID{ID: "cert-2"}},
		},
	}
	err := RemapRefs(data, map[string]string{"cert-1": "new-1", "cert-2": "new-2"})
	assert.NoError(t, err)
	assert.Equal(t, "new-1", data.Domains[0].SSLCert.ID)
	assert.Equal(t, "new-2", data.Domains[1].SSLCert.ID)
}

// An id with no mapping entry is left alone rather than blanked.
func TestRemapRefsLeavesUnmappedReferencesUntouched(t *testing.T) {
	data := &SSLCert{Provider: ObjectID{ID: "keep-me"}}
	assert.NoError(t, RemapRefs(data, map[string]string{"other": "new"}))
	assert.Equal(t, "keep-me", data.Provider.ID)
}

// The self-check is the whole safety property: if the walk cannot reach a
// reference, RemapRefs must say so rather than return a half-rewritten value
// that imports as a pointer into another installation's data.
func TestRemapRefsSelfCheckDetectsAMissedReference(t *testing.T) {
	data := &unreachableRefData{hidden: ObjectID{ID: "old"}}
	err := RemapRefs(data, map[string]string{"old": "new"})

	// Error() renders the code; the missed identifier goes to the debug log via
	// WithMsgLog, which is this codebase's convention for detail an end user
	// should not see. What matters here is that it fails at all rather than
	// returning a half-rewritten value.
	assert.Error(t, err)
	assert.ErrorIs(t, err, hperrors.ErrInternal)

	// And the reference really was left unrewritten, which is the condition the
	// self-check exists to catch.
	assert.Equal(t, "old", data.hidden.ID)
}

func TestRemapRefsIsANoOpWithoutAMapping(t *testing.T) {
	data := &SSLCert{Provider: ObjectID{ID: "p1"}}
	assert.NoError(t, RemapRefs(data, nil))
	assert.Equal(t, "p1", data.Provider.ID)
}

// Encrypted values must never be walked into: their contents are unexported and
// a stray rewrite there would corrupt a secret beyond recovery.
func TestRemapRefsDoesNotDisturbEncryptedFields(t *testing.T) {
	useDataKey(t)
	data := &Secret{Key: "DB", Value: NewEncryptedField("old")}
	assert.NoError(t, RemapRefs(data, map[string]string{"old": "new"}))

	plain, err := data.Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "old", plain, "a secret that happens to equal a ref id must not move")
}

// unreachableRefData reports a reference held in an unexported field, which no
// reflection walk can write. It exists only to prove the self-check fires.
type unreachableRefData struct {
	hidden ObjectID
}

func (d *unreachableRefData) GetType() base.SettingType { return base.SettingTypeScript }
func (d *unreachableRefData) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{RefSettingIDs: []string{d.hidden.ID}}
}
func (d *unreachableRefData) GetResourceLinks(*Setting) []*ResLink { return nil }
func (d *unreachableRefData) Migrate(*Setting) (bool, error)       { return false, nil }

// The premise of the value-driven walk, asserted rather than assumed: setting
// types really do hold references in shapes other than ObjectID, so a walk that
// recognized only ObjectID would miss them. If this ever stops being true the
// design could be simplified - until then it must not be.
func TestReferencesAreHeldInShapesOtherThanObjectID(t *testing.T) {
	objectIDType := reflect.TypeFor[ObjectID]()

	backup := reflect.TypeFor[SystemBackup]()
	cloudStorage, ok := backup.FieldByName("CloudStorage")
	assert.True(t, ok)
	assert.NotEqual(t, objectIDType, cloudStorage.Type,
		"SystemBackup.CloudStorage holds a reference but is not an ObjectID")

	idField, ok := cloudStorage.Type.FieldByName("ID")
	assert.True(t, ok, "and it carries the reference in a bare ID string")
	assert.Equal(t, reflect.String, idField.Type.Kind())
}
