package settingeventserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

type fakeMounts struct {
	settingmountservice.Service
	recorded []*entity.Setting
}

func (f *fakeMounts) RecordRefresh(
	_ context.Context, _ database.IDB, settings ...*entity.Setting,
) (*entity.Task, error) {
	f.recorded = append(f.recorded, settings...)
	return &entity.Task{ID: "task_1"}, nil
}

// Every path the settings screens write through records a refresh for what it
// wrote, and hands the task back to be scheduled after commit.
func TestSettingEventsRecordAMountRefresh(t *testing.T) {
	mounts := &fakeMounts{}
	s := &service{settingMountService: mounts}
	cert := &entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert}
	ctx := context.Background()

	created := &settingeventservice.CreateEvent{Setting: cert}
	updated := &settingeventservice.UpdateEvent{Setting: cert}
	status := &settingeventservice.UpdateEvent{Setting: cert}
	deleted := &settingeventservice.DeleteEvent{Setting: cert}
	assert.NoError(t, s.OnCreate(ctx, nil, created))
	assert.NoError(t, s.OnUpdate(ctx, nil, updated))
	assert.NoError(t, s.OnUpdateStatus(ctx, nil, status))
	assert.NoError(t, s.OnDelete(ctx, nil, deleted))

	assert.Len(t, mounts.recorded, 4)
	for _, tasks := range [][]*entity.Task{created.Tasks, updated.Tasks, status.Tasks, deleted.Tasks} {
		assert.Len(t, tasks, 1)
	}
}
