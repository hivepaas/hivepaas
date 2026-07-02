package appcopyserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) applySchedJobSettings(
	ctx context.Context,
	db database.Tx,
	data *appCopyData,
) error {
	app := data.TargetApp
	jobSettings := app.GetSettingsByType(base.SettingTypeSchedJob)

	err := s.taskQueue.ScheduleTasksForSchedJobs(ctx, db, jobSettings, false)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
