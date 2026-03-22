package taskcronjobexec

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type sysCleanupTaskData struct {
	*taskData
	TaskOutput *entity.TaskSystemCleanupOutput
}

func (e *Executor) cronExecSystemCleanup(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) error {
	setting := data.RefObjects.RefSettings[data.CronJob.TargetSetting.ID]
	if setting == nil {
		return apperrors.NewNotFound("System cleanup settings")
	}
	sysCleanup := setting.MustAsSystemCleanup()

	taskData := &sysCleanupTaskData{
		taskData: data,
		TaskOutput: &entity.TaskSystemCleanupOutput{
			DBCleanup:      &entity.DBCleanupOutput{},
			ClusterCleanup: &entity.ClusterCleanupOutput{},
		},
	}

	// Cleanup system DB
	err1 := e.sysDBCleanup(ctx, db, sysCleanup.DBObjectRetention, taskData)

	// Cleanup unused cluster data (docker)
	err2 := e.sysClusterCleanup(ctx, sysCleanup.ClusterCleanup, taskData)

	// Assign back the result output
	data.Task.MustSetOutput(taskData.TaskOutput)

	return errors.Join(err1, err2)
}
