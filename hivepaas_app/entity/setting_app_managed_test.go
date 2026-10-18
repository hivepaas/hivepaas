package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func managedPostgres() *AppManagedSettings {
	return &AppManagedSettings{
		Engine: "postgres", EngineAPI: 1,
		Postgres: &ManagedPostgres{
			Storage: ManagedStorage{BackupRepo: ObjectID{ID: "repo-1"}, KeyAuth: ObjectID{ID: "key-1"},
				Prefix: "backups/walg/shop-db"},
			EncryptionKey:  NewEncryptedField("0123456789abcdef0123456789abcdef"),
			ArchiveTimeout: timeutil.Duration(time.Minute),
			RetainFull:     7,
			RestoreTest: ManagedRestoreTest{Mode: ManagedRestoreTestVerifier,
				VerifierApp: ObjectID{ID: "app-verifier"}, CheckQuery: "SELECT 1"},
			Jobs: ManagedPostgresJobs{Nightly: ObjectID{ID: "job-1"}, Health: ObjectID{ID: "job-2"},
				RestoreTest: ObjectID{ID: "job-3"}},
		},
	}
}

func TestAppManagedSettingsRoundTrip(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	setting := &Setting{Type: base.SettingTypeAppManaged, Version: CurrentAppManagedVersion}
	setting.MustSetData(managedPostgres())

	stored := &Setting{Type: base.SettingTypeAppManaged, Version: setting.Version, Data: setting.Data}
	managed, err := stored.AsAppManagedSettings()

	assert.NoError(t, err)
	assert.Equal(t, "postgres", managed.Engine)
	assert.Equal(t, "backups/walg/shop-db", managed.Postgres.Storage.Prefix)
	assert.NotContains(t, setting.Data, "0123456789abcdef", "the WAL-G key is stored encrypted")
	assert.NoError(t, managed.Decrypt())
	plain, err := managed.Postgres.EncryptionKey.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", plain)
}

// What the engine depends on cannot be deleted under it: the repository, the
// storage and key, the verifier app. Its jobs point at it, not the other way
// round, so that removing the whole is not a cycle.
func TestAppManagedSettingsReferences(t *testing.T) {
	refs := managedPostgres().GetRefObjectIDs()

	assert.ElementsMatch(t, []string{"repo-1", "key-1"}, refs.RefSettingIDs)
	assert.Equal(t, []string{"app-verifier"}, refs.RefAppIDs)
}

func TestAManagedJobPointsAtItsOwner(t *testing.T) {
	job := &SchedJob{JobType: base.SchedJobTypeContainerCommand, App: ObjectID{ID: "app-1"},
		ManagedBy: ObjectID{ID: "managed-1"}}

	assert.True(t, job.IsManaged())
	assert.Contains(t, job.GetRefObjectIDs().RefSettingIDs, "managed-1")
	assert.False(t, (&SchedJob{}).IsManaged())
}
