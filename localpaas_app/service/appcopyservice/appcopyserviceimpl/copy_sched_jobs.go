package appcopyserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
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
		return apperrors.Wrap(err)
	}

	return nil
}
