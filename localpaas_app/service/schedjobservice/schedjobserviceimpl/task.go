package schedjobserviceimpl

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

func (s *service) CreateSchedJobTask(
	jobSetting *entity.Setting,
	runAt time.Time,
	timeNow time.Time,
) (*entity.Task, error) {
	schedJob, err := jobSetting.AsSchedJob()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		TargetID: jobSetting.ID,
		Type:     base.TaskTypeSchedJobExec,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:        schedJob.Priority,
			MaxRetry:        schedJob.MaxRetry,
			RetryDelay:      schedJob.RetryDelay,
			Timeout:         schedJob.Timeout,
			ControlDisabled: schedJob.ControlDisabled,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     runAt,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}, nil
}
