package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func setting(id string, typ base.SettingType, data entity.SettingData) *entity.Setting {
	s := &entity.Setting{ID: id, Type: typ, Status: base.SettingStatusActive, UpdateVer: 3}
	s.MustSetData(data)
	return s
}

func managedApp() *entity.App {
	return &entity.App{ID: "app-1", Settings: []*entity.Setting{
		setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres"}),
		setting("job-1", base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			ManagedBy: entity.ObjectID{ID: "managed-1"}}),
		setting("job-2", base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			ManagedBy: entity.ObjectID{ID: "managed-1"}}),
		setting("job-mine", base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand}),
		setting("artifacts-1", base.SettingTypeAppArtifacts, &entity.AppArtifactsSettings{}),
		setting("secret-1", base.SettingTypeSecret, &entity.Secret{Key: "WALG_LIBSODIUM_KEY"}),
	}}
}

// Unmanaging removes the engine's setting and the jobs it made, and nothing the
// database needs to go on archiving: its artifacts, its secrets, its arguments
// stay, so a person takes over a database that still works.
func TestUnmanageRemovesTheEngineAndItsJobsOnly(t *testing.T) {
	now := time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)

	deleting, err := unmanagedSettings(managedApp(), now)

	assert.NoError(t, err)
	ids := make([]string, 0, len(deleting))
	for _, s := range deleting {
		ids = append(ids, s.ID)
		assert.Equal(t, now, s.DeletedAt, s.ID)
		assert.Equal(t, 4, s.UpdateVer, s.ID)
	}
	assert.ElementsMatch(t, []string{"managed-1", "job-1", "job-2"}, ids)
}

func TestUnmanagingAnOrdinaryAppIsRefused(t *testing.T) {
	_, err := unmanagedSettings(&entity.App{ID: "app-1"}, time.Now())

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestUnmanagingFindsTheVerifierToDelete(t *testing.T) {
	managed := setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres",
		Database: &entity.ManagedDatabase{RestoreTest: entity.ManagedRestoreTest{
			Mode: entity.ManagedRestoreTestVerifier, VerifierApp: entity.ObjectID{ID: "ver-1"}}}})

	assert.Equal(t, "ver-1", verifierOf(managed))
	assert.Empty(t, verifierOf(managedApp().Settings[0]), "a test in the container has none")
}
