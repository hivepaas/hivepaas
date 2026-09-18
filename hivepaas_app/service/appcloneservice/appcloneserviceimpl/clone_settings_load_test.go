package appcloneserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeCloneAppService struct {
	appservice.Service
	app *entity.App
}

func (f *fakeCloneAppService) LoadApp(
	_ context.Context, _ database.IDB, _, _ string, _, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	return f.app, nil
}

type fakeCloneSettingService struct {
	settingservice.Service
}

func (f *fakeCloneSettingService) LoadRefObjectsByIDs(
	context.Context, database.IDB, **entity.RefObjects, *entity.ObjectScope, bool, *entity.RefObjectIDs,
) error {
	return nil
}

// srcAppWithCloneSettings is what the task path loads: the app, carrying the
// clone settings somebody saved for it.
func srcAppWithCloneSettings(t *testing.T, settings *entity.AppCloneSettings) *entity.App {
	t.Helper()
	app := &entity.App{ID: "src-1", Name: "blog", Key: "blog", ProjectID: "p1"}
	if settings == nil {
		return app
	}
	setting := &entity.Setting{ID: "set-clone", Type: base.SettingTypeAppClone, ObjectID: app.ID}
	assert.NoError(t, setting.SetData(settings))
	app.Settings = []*entity.Setting{setting}
	return app
}

func cloneTask(t *testing.T) *entity.Task {
	t.Helper()
	task := &entity.Task{ID: "task-1", Type: base.TaskTypeAppClone}
	task.MustSetArgs(&entity.TaskAppCloneArgs{SrcApp: entity.ObjectID{ID: "src-1"}})
	return task
}

func loadCloneData(t *testing.T, app *entity.App, req *appcloneservice.AppCloneReq) (*appCloneData, error) {
	t.Helper()
	svc := &service{
		appService:     &fakeCloneAppService{app: app},
		settingService: &fakeCloneSettingService{},
	}
	data := &appCloneData{AppCloneReq: req}
	return data, svc.loadAppCloneData(context.Background(), nil, data)
}

// The task carries nothing but the source app's id: what to clone is the
// settings saved against that app, and the request that scheduled the task
// already checked the name in them was free.
func TestLoadAppCloneDataReadsTheSettingsTheTaskWasScheduledFor(t *testing.T) {
	saved := &entity.AppCloneSettings{
		TargetName: "blog-staging", TargetEnv: "staging",
		CloneEnvVars: true, CloneSecrets: true, CloneVolumes: true,
	}
	req := &appcloneservice.AppCloneReq{TaskExecData: &queue.TaskExecData{Task: cloneTask(t)}}

	data, err := loadCloneData(t, srcAppWithCloneSettings(t, saved), req)

	assert.NoError(t, err)
	assert.Equal(t, "blog-staging", data.CloneSettings.TargetName)
	assert.Equal(t, "staging", data.CloneSettings.TargetEnv)
	assert.True(t, data.CloneSettings.CloneEnvVars)
	assert.True(t, data.CloneSettings.CloneSecrets)
	assert.True(t, data.CloneSettings.CloneVolumes)
}

// Cloning nothing is not a safe default here: the app it would create is not the
// app that was asked for, and it would take the name of one that was.
func TestLoadAppCloneDataRefusesATaskWhoseSettingsAreGone(t *testing.T) {
	req := &appcloneservice.AppCloneReq{TaskExecData: &queue.TaskExecData{Task: cloneTask(t)}}

	_, err := loadCloneData(t, srcAppWithCloneSettings(t, nil), req)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

// Creating a preview app clones through this with its own source app and its own
// callbacks, and must not pick up whatever the app was last cloned with.
func TestLoadAppCloneDataLeavesACallersOwnSourceAppAlone(t *testing.T) {
	saved := &entity.AppCloneSettings{TargetName: "blog-staging", CloneVolumes: true}
	app := srcAppWithCloneSettings(t, saved)
	req := &appcloneservice.AppCloneReq{
		TaskExecData: &queue.TaskExecData{Task: cloneTask(t)},
		SrcApp:       app,
	}

	data, err := loadCloneData(t, app, req)

	assert.NoError(t, err)
	assert.Empty(t, data.CloneSettings.TargetName)
	assert.False(t, data.CloneSettings.CloneVolumes)
}
