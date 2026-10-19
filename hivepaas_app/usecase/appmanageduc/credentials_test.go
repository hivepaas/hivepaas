package appmanageduc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

func secretSetting(t *testing.T, id, key, value string) *entity.Setting {
	t.Helper()
	s := &entity.Setting{ID: id, Type: base.SettingTypeSecret, Name: key, UpdateVer: 1}
	s.MustSetData(&entity.Secret{Key: key, Value: entity.NewEncryptedField(value)})
	return s
}

// Refreshing rewrites the secrets whose value the storage's key changed, and
// only those: one that did not change is not mounted again.
func TestRefreshRewritesTheSecretsThatChanged(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	access := secretSetting(t, "s1", "WALG_AWS_ACCESS_KEY_ID", "AKIA-OLD")
	secret := secretSetting(t, "s2", "WALG_AWS_SECRET_ACCESS_KEY", "same-secret")
	sodium := secretSetting(t, "s3", "WALG_LIBSODIUM_KEY", "never-touched")

	changed, err := refreshedSecrets([]*entity.Setting{access, secret, sodium}, map[string]string{
		"WALG_AWS_ACCESS_KEY_ID": "AKIA-NEW", "WALG_AWS_SECRET_ACCESS_KEY": "same-secret"})

	assert.NoError(t, err)
	assert.Equal(t, []*entity.Setting{access}, changed)
	assert.Equal(t, 2, access.UpdateVer)
	stored := &entity.Setting{Type: base.SettingTypeSecret, Data: access.Data}
	data, err := stored.AsSecret()
	assert.NoError(t, err)
	plain, _ := data.Value.GetPlain()
	assert.Equal(t, "AKIA-NEW", plain)
	assert.Equal(t, 1, sodium.UpdateVer)
}

// The verifier's copies are refreshed too, under its own narrower key.
func TestTheVerifierReadsTheArchiveWithItsOwnKey(t *testing.T) {
	pg := &entity.ManagedDatabase{Storage: entity.ManagedStorage{CloudStorage: entity.ObjectID{ID: "s3-1"},
		KeyAuth: entity.ObjectID{ID: "key-rw"}, Prefix: "walg/db"}}
	app := &entity.App{ID: "db-1"}

	assert.Equal(t, []archiveReader{{appID: "db-1", storage: pg.Storage}}, archiveReaders(app, pg))

	pg.RestoreTest = entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestVerifier,
		VerifierApp: entity.ObjectID{ID: "ver-1"}, KeyAuth: entity.ObjectID{ID: "key-ro"}}
	readers := archiveReaders(app, pg)
	if assert.Len(t, readers, 2) {
		assert.Equal(t, "ver-1", readers[1].appID)
		assert.Equal(t, "key-ro", readers[1].storage.KeyAuth.ID)
		assert.Equal(t, "walg/db", readers[1].storage.Prefix, "the database's archive")
		assert.Equal(t, "key-rw", readers[0].storage.KeyAuth.ID, "the database keeps its own")
	}

	pg.RestoreTest.KeyAuth = entity.ObjectID{}
	assert.Equal(t, "key-rw", archiveReaders(app, pg)[1].storage.KeyAuth.ID, "without one, the database's")
}
