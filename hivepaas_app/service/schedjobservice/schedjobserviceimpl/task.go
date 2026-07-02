package schedjobserviceimpl

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) CreateSchedJobTask(
	jobSetting *entity.Setting,
	runAt time.Time,
	timeNow time.Time,
) (*entity.Task, error) {
	schedJob, err := jobSetting.AsSchedJob()
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    jobSetting.Scope,
		ObjectID: jobSetting.ObjectID,
		TargetID: jobSetting.ID,
		Type:     base.TaskTypeSchedJobExec,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:           schedJob.Priority,
			MaxRetry:           schedJob.MaxRetry,
			RetryDelay:         schedJob.RetryDelay,
			RetryDelayIncr:     schedJob.RetryDelayIncr,
			RetryBackoffJitter: schedJob.RetryBackoffJitter,
			RetryDelayMax:      schedJob.RetryDelayMax,
			Timeout:            schedJob.Timeout,
			ControlDisabled:    schedJob.ControlDisabled,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     runAt,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}, nil
}
