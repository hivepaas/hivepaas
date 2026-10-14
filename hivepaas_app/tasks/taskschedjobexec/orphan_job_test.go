package taskschedjobexec

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// goneScope finds no object for any scope: it was deleted.
type goneScope struct{ scopeservice.Service }

func (goneScope) LoadObjectScope(context.Context, database.IDB, base.ObjectScopeType, string,
	bool) (*entity.ObjectScope, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppNotFound)
}

// deletedSettings keeps which objects' settings were deleted.
type deletedSettings struct {
	repository.SettingRepo
	scope     base.ObjectScopeType
	objectIDs []string
}

func (r *deletedSettings) DeleteAllByObjects(_ context.Context, _ database.IDB, scope base.ObjectScopeType,
	objectIDs []string, _ ...bunex.DeleteQueryOption) error {
	r.scope, r.objectIDs = scope, objectIDs
	return nil
}

// A job saved while its app was being deleted outlives the app. Its run finds
// the app gone: the run fails, without retries, and the app's settings - the
// job among them - are deleted, as the app's deletion would have.
func TestARunOfAJobWhoseAppIsGoneDeletesTheJob(t *testing.T) {
	settings := &deletedSettings{}
	e := &Executor{scopeService: goneScope{}, settingRepo: settings}
	data := &taskData{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{}},
		SchedJob:     &entity.Setting{Scope: base.ObjectScopeApp, ObjectID: "app"},
	}

	err := e.loadSchedJobData(context.Background(), database.Tx{}, data)

	assert.True(t, errors.Is(err, hperrors.ErrNotFound), "got %v", err)
	assert.True(t, data.TaskNonRetryable, "the object will not come back")
	assert.Equal(t, base.ObjectScopeApp, settings.scope)
	assert.Equal(t, []string{"app"}, settings.objectIDs)
	assert.Nil(t, data.Scope)
}

// A run that ended without its scope has nobody to tell how it went; it used
// to take the process down, building the message from a nil scope.
func TestARunWithoutItsScopeSendsNoNotification(t *testing.T) {
	e := &Executor{} // a notification would reach the nil notification service
	data := &taskData{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{Status: base.TaskStatusFailed}},
		SchedJob:     &entity.Setting{Scope: base.ObjectScopeApp, ObjectID: "app"},
	}
	assert.True(t, data.Task.IsFailedCompletely(), "a run that would be notified")

	assert.NotPanics(t, func() { e.onPostTx(context.Background(), data) })
}
