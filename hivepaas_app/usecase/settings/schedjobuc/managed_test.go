package schedjobuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func jobSetting(t *testing.T, job *entity.SchedJob) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{Type: base.SettingTypeSchedJob, Version: entity.CurrentSchedJobVersion}
	setting.MustSetData(job)
	return setting
}

// A job its app's engine made is changed through the engine's settings: the
// generic screens refuse to edit, pause or delete it, and name what they refused.
func TestAManagedJobIsRefusedHere(t *testing.T) {
	managed := jobSetting(t, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
		ManagedBy: entity.ObjectID{ID: "managed-1"}})

	err := refuseManagedJob(managed)

	assert.ErrorIs(t, err, hperrors.ErrAppManagedFieldLocked)
	info, _ := hperrors.ParseError(err, "en")
	assert.Equal(t, map[string]any{"Field": "schedJob"}, info.Params)
}

func TestAnOrdinaryJobIsNotRefused(t *testing.T) {
	assert.NoError(t, refuseManagedJob(jobSetting(t, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand})))
}
