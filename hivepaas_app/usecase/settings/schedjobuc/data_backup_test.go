package schedjobuc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestDataBackupsLiveInApps(t *testing.T) {
	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeApp, base.SchedJobTypeDataBackup))
	for _, scope := range []base.ObjectScopeType{base.ObjectScopeGlobal, base.ObjectScopeProject,
		base.ObjectScopeProjectEnv} {
		err := checkJobTypeInScope(scope, base.SchedJobTypeDataBackup)
		assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "%s: got %v", scope, err)
	}
}

// A data backup is its own app's: the command runs there, the volume is one it
// mounts.
func TestADataBackupIsItsAppsOwn(t *testing.T) {
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app-1"}
	job := &entity.SchedJob{JobType: base.SchedJobTypeDataBackup, App: entity.ObjectID{ID: "app-1"},
		DataBackup: &entity.SchedJobDataBackup{Source: base.SchedJobDataBackupSourceCommand}}
	assert.NoError(t, checkDataBackupApp(scope, job))

	job.App.ID = "app-2"
	err := checkDataBackupApp(scope, job)
	assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "got %v", err)

	assert.NoError(t, checkDataBackupApp(scope, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand}))
}
