package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) applySchedJobSettings(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	app := data.DestApp
	jobSettings := app.GetSettingsByType(base.SettingTypeSchedJob)

	tx, ok := db.(database.Tx)
	if !ok {
		return apperrors.Wrap(apperrors.ErrInternal).WithMsgLog("db is not transaction")
	}

	err := s.taskQueue.ScheduleTasksForSchedJobs(ctx, tx, jobSettings, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
