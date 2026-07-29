package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) SetProjectEnvStatus(
	ctx context.Context,
	db database.IDB,
	projectEnv *entity.ProjectEnv,
	status base.ProjectStatus,
	recursive bool,
) error {
	var targetAppStatus base.AppStatus
	switch projectEnv.Status {
	case base.ProjectStatusActive:
		targetAppStatus = base.AppStatusActive
	case base.ProjectStatusDisabled:
		targetAppStatus = base.AppStatusDisabled
	case base.ProjectStatusDeleting:
		// Do nothing
	}

	for _, app := range projectEnv.Apps {
		if targetAppStatus == "" {
			continue
		}
		app.Project = projectEnv.Project
		app.ProjectEnv = projectEnv
		// Run app update in a separate transaction to reduce lock time
		err := s.ExecuteEnvInTx(ctx, projectEnv, true, func(db database.Tx) error {
			err := s.appService.SetAppStatus(ctx, db, app, targetAppStatus, recursive)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}

	if projectEnv.Status == status {
		return nil
	}
	projectEnv.Status = status
	projectEnv.UpdatedAt = timeutil.NowUTC()
	projectEnv.UpdateVer++

	err := s.projectEnvRepo.Update(ctx, db, projectEnv,
		bunex.UpdateColumns("status", "updated_at", "update_ver"))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
