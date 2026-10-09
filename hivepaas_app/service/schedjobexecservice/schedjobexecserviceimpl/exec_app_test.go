package schedjobexecserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

// A job whose app is not there - none named, or one deleted since - fails
// saying so, rather than in a panic.
func TestARunWithoutItsAppFailsSayingSo(t *testing.T) {
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob}
	setting.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
		App: entity.ObjectID{ID: "gone"}, Command: &entity.CommandTemplate{Command: "echo hi"}})

	_, err := (&service{}).SchedJobExec(context.Background(), database.Tx{},
		&schedjobexecservice.SchedJobExecReq{SchedJobSetting: setting})

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}
