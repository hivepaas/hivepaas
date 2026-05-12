package cronjobserviceimpl

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

func (s *service) CreateCronJobTask(
	jobSetting *entity.Setting,
	runAt time.Time,
	timeNow time.Time,
) (*entity.Task, error) {
	cronJob, err := jobSetting.AsCronJob()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		TargetID: jobSetting.ID,
		Type:     base.TaskTypeCronJobExec,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:        cronJob.Priority,
			MaxRetry:        cronJob.MaxRetry,
			RetryDelay:      cronJob.RetryDelay,
			Timeout:         cronJob.Timeout,
			ControlDisabled: cronJob.ControlDisabled,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     runAt,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}, nil
}
