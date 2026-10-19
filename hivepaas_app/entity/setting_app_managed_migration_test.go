package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A database made before engines were named apart kept its state under
// "postgres", and its points' and restore's positions as "lsn". The migration
// moves both to the engine-neutral names, and the row is written back without
// the old ones.
func TestAppManagedMigrationMovesPostgresToDatabase(t *testing.T) {
	stored := &entity.Setting{
		ID: "s1", Type: base.SettingTypeAppManaged, Version: 1,
		Data: `{
			"engine": "postgres", "engineApi": 1,
			"locks": {"image": "postgres:18"},
			"postgres": {
				"storage": {"prefix": "backups/walg/shop-db", "placement": "off-site"},
				"retainFull": 7,
				"restorePoints": [{"name": "before-v42", "lsn": "0/3000028", "at": "2026-10-18T10:00:00Z"}],
				"restore": {"sourceApp": {"id": "app-0"},
					"target": {"kind": "point", "point": "before-v42", "lsn": "0/3000028"}}
			}
		}`,
	}

	changed, err := stored.Migrate()

	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, entity.CurrentAppManagedVersion, stored.Version)
	assert.Contains(t, stored.Data, `"database"`)
	assert.NotContains(t, stored.Data, `"postgres":`)
	assert.NotContains(t, stored.Data, `"lsn"`)

	stored.ResetParsedData()
	managed := stored.MustAsAppManagedSettings()
	if assert.NotNil(t, managed.Database) {
		assert.Equal(t, "backups/walg/shop-db", managed.Database.Storage.Prefix)
		assert.Equal(t, 7, managed.Database.RetainFull)
		assert.Equal(t, "0/3000028", managed.Database.RestorePoint("before-v42").Position)
		if assert.NotNil(t, managed.Database.Restore) {
			assert.Equal(t, "0/3000028", managed.Database.Restore.Target.Position)
		}
	}
	assert.Nil(t, managed.LegacyPostgres)
	assert.Equal(t, "postgres:18", managed.Locks.Image)
}

// Startup data migration does not reach every row: a row still in version 1's
// shape is read in the new one all the same, and written so, at the current
// version, by the first save.
func TestAnAppManagedRowOfVersion1IsReadInTheNewShape(t *testing.T) {
	stored := &entity.Setting{
		ID: "s1", Type: base.SettingTypeAppManaged, Version: 1,
		Data: `{"engine": "postgres", "engineApi": 1, "locks": {},
			"postgres": {"storage": {"prefix": "p"}, "restorePoints": [{"name": "x", "lsn": "0/1"}],
				"restore": {"sourceApp": {"id": "app-0"}, "target": {"kind": "point", "point": "x", "lsn": "0/1"}}}}`,
	}

	managed := stored.MustAsAppManagedSettings()

	if assert.NotNil(t, managed.Database) {
		assert.Equal(t, "p", managed.Database.Storage.Prefix)
		assert.Equal(t, "0/1", managed.Database.RestorePoint("x").Position)
		assert.Equal(t, "0/1", managed.Database.Restore.Target.Position)
	}
	assert.NoError(t, stored.SetAppManagedData(managed))
	assert.Equal(t, entity.CurrentAppManagedVersion, stored.Version)
	assert.Contains(t, stored.Data, `"database"`)
	assert.NotContains(t, stored.Data, `"postgres":`)
	assert.NotContains(t, stored.Data, `"lsn"`)
}

// A member keeps no engine state: there is nothing to move, and the row only
// takes the new version.
func TestAppManagedMigrationOfAMember(t *testing.T) {
	stored := &entity.Setting{
		ID: "s1", Type: base.SettingTypeAppManaged, Version: 1,
		Data: `{"engine": "postgres", "engineApi": 1, "locks": {}, "memberOf": {"app": {"id": "app-1"},
			"role": "replica"}}`,
	}

	changed, err := stored.Migrate()

	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, entity.CurrentAppManagedVersion, stored.Version)
	managed := stored.MustAsAppManagedSettings()
	assert.Nil(t, managed.Database)
	assert.Equal(t, "app-1", managed.MemberOf.App.ID)
}

func TestAppManagedMigrationLeavesTheCurrentVersionAlone(t *testing.T) {
	data := `{"engine":"postgres","engineApi":1,"locks":{},"database":{"storage":{"prefix":"p"},` +
		`"restorePoints":[{"name":"x","position":"0/1","at":"2026-10-18T10:00:00Z"}]}}`
	current := &entity.Setting{ID: "s1", Type: base.SettingTypeAppManaged,
		Version: entity.CurrentAppManagedVersion, Data: data}

	changed, err := current.Migrate()

	assert.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, data, current.Data)
	assert.Equal(t, "0/1", current.MustAsAppManagedSettings().Database.RestorePoint("x").Position)

	newer := &entity.Setting{Type: base.SettingTypeAppManaged, Version: entity.CurrentAppManagedVersion + 1,
		Data: "{}"}
	_, err = newer.Migrate()
	assert.ErrorIs(t, err, hperrors.ErrDataVerNewerThanSystemVer)
}
