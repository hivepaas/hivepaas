package tasksettingmountrefresh

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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeApps struct{ repository.AppRepo }

func (fakeApps) GetByID(
	_ context.Context, _ database.IDB, _, id string, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	if id == "app_gone" {
		return nil, hperrors.NewNotFound("App")
	}
	return &entity.App{ID: id}, nil
}

type fakeMounts struct {
	settingmountservice.Service
	refreshed []string
}

func (f *fakeMounts) Refresh(_ context.Context, _ database.IDB, app *entity.App) error {
	f.refreshed = append(f.refreshed, app.ID)
	if app.ID == "app_broken" {
		return errors.New("docker said no")
	}
	return nil
}

// Every app is tried: one that fails is written down and retried with the task,
// one deleted since is skipped.
func TestTheTaskRefreshesEveryAppAndReportsThoseThatFailed(t *testing.T) {
	mounts := &fakeMounts{}
	e := &Executor{appRepo: fakeApps{}, settingMountService: mounts, logger: logging.GlobalLogger(),
		listPreviews: noPreviews}
	task := &entity.Task{Type: base.TaskTypeSettingMountRefresh}
	assert.NoError(t, task.SetArgs(&entity.TaskSettingMountRefreshArgs{
		AppIDs: []string{"app_1", "app_broken", "app_gone", "app_2"}}))

	err := e.execute(context.Background(), database.Tx{}, &queue.TaskExecData{Task: task})

	assert.Error(t, err)
	assert.Equal(t, []string{"app_1", "app_broken", "app_2"}, mounts.refreshed)
	out, _ := task.OutputAsSettingMountRefresh()
	assert.Equal(t, 2, out.Applied)
	assert.Contains(t, out.Failed, "app_broken")
}

func noPreviews(context.Context, database.IDB, *entity.App) ([]*entity.App, error) {
	return nil, nil
}

// A preview mounts its parent's inheritable entries, so it is refreshed with it.
func TestTheTaskRefreshesAnAppsPreviewsWithIt(t *testing.T) {
	mounts := &fakeMounts{}
	e := &Executor{appRepo: fakeApps{}, settingMountService: mounts, logger: logging.GlobalLogger()}
	e.listPreviews = func(_ context.Context, _ database.IDB, app *entity.App) ([]*entity.App, error) {
		if app.ID == "app_1" {
			return []*entity.App{{ID: "app_1_preview", ParentID: "app_1"}}, nil
		}
		return nil, nil
	}
	task := &entity.Task{Type: base.TaskTypeSettingMountRefresh}
	assert.NoError(t, task.SetArgs(&entity.TaskSettingMountRefreshArgs{AppIDs: []string{"app_1", "app_2"}}))

	assert.NoError(t, e.execute(context.Background(), database.Tx{}, &queue.TaskExecData{Task: task}))

	assert.Equal(t, []string{"app_1", "app_1_preview", "app_2"}, mounts.refreshed)
	out, _ := task.OutputAsSettingMountRefresh()
	assert.Equal(t, 3, out.Applied)
}
