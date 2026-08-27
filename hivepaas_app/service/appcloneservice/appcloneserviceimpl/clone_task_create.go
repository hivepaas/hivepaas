package appcloneserviceimpl

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) CreateAppCloneTask(
	app *entity.App,
) (*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	appCloneTask := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Type:     base.TaskTypeAppClone,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority: base.TaskPriorityDefault,
			Timeout:  timeutil.Duration(base.AppCloneTimeoutDefault),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err := appCloneTask.SetArgs(&entity.TaskAppCloneArgs{
		SrcApp: entity.ObjectID{ID: app.ID},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return appCloneTask, nil
}
