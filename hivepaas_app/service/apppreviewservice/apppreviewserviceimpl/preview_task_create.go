package apppreviewserviceimpl

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) CreateAppPreviewTask(
	app *entity.App,
	args *entity.TaskAppPreviewArgs,
) (*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	if args == nil {
		args = &entity.TaskAppPreviewArgs{}
	}
	args.ParentApp = entity.ObjectID{ID: app.ID}

	appPreviewTask := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Type:     base.TaskTypeAppPreview,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority: base.TaskPriorityDefault,
			Timeout:  timeutil.Duration(base.AppPreviewTimeoutDefault),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err := appPreviewTask.SetArgs(args)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return appPreviewTask, nil
}
