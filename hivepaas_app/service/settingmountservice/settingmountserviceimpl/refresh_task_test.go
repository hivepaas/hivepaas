package settingmountserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type fakeTaskRepo struct {
	repository.TaskRepo
	inserted []*entity.Task
}

func (f *fakeTaskRepo) Insert(
	_ context.Context, _ database.IDB, task *entity.Task, _ ...bunex.InsertQueryOption,
) error {
	f.inserted = append(f.inserted, task)
	return nil
}

func recorder(t *testing.T, readers map[string][]string) (*service, *fakeTaskRepo) {
	t.Helper()
	tasks := &fakeTaskRepo{}
	svc := &service{taskRepo: tasks}
	svc.loadReaders = func(_ context.Context, _ database.IDB, ids []string) ([]string, error) {
		var apps []string
		for _, id := range ids {
			apps = append(apps, readers[id]...)
		}
		return apps, nil
	}
	return svc, tasks
}

func TestRecordRefreshNamesTheAppsThatReadTheSettings(t *testing.T) {
	svc, tasks := recorder(t, map[string][]string{"cert_1": {"app_2", "app_1"}})

	task, err := svc.RecordRefresh(context.Background(), nil,
		&entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert},
		&entity.Setting{ID: "entry_1", Type: base.SettingTypeAppSettingMount, ObjectID: "app_1"},
		&entity.Setting{ID: "secret_1", Type: base.SettingTypeSecret, ObjectID: "app_9"})

	assert.NoError(t, err)
	if assert.NotNil(t, task) && assert.Len(t, tasks.inserted, 1) {
		assert.Equal(t, base.TaskTypeSettingMountRefresh, task.Type)
		assert.Equal(t, base.TaskStatusNotStarted, task.Status)
		args, err := task.ArgsAsSettingMountRefresh()
		assert.NoError(t, err)
		assert.Equal(t, []string{"app_1", "app_2"}, args.AppIDs, "sorted, once each")
	}
}

func TestRecordRefreshRecordsNothingWhenNoAppReads(t *testing.T) {
	svc, tasks := recorder(t, nil)

	task, err := svc.RecordRefresh(context.Background(), nil,
		&entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert},
		&entity.Setting{ID: "secret_1", Type: base.SettingTypeSecret})

	assert.NoError(t, err)
	assert.Nil(t, task)
	assert.Empty(t, tasks.inserted)
}
