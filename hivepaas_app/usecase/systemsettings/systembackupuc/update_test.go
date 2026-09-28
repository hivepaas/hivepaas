package systembackupuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A system backup goes into a repository of the global scope: the installation's
// own backup is no project's.
func TestCheckTargetRepository(t *testing.T) {
	global := &entity.Setting{ID: "r1", Type: base.SettingTypeBackupRepo, Scope: base.ObjectScopeGlobal,
		Status: base.SettingStatusActive}
	assert.NoError(t, checkTargetRepository(global))

	project := &entity.Setting{ID: "r2", Type: base.SettingTypeBackupRepo, Scope: base.ObjectScopeProject,
		ObjectID: "p1", Status: base.SettingStatusActive}
	assert.ErrorIs(t, checkTargetRepository(project), hperrors.ErrArgumentInvalid)

	assert.ErrorIs(t, checkTargetRepository(nil), hperrors.ErrNotFound)
}
