package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
) (changedApps []*entity.App, err error) {
	if projectEnv.Status == status {
		return nil, nil
	}

	var targetAppStatus base.AppStatus
	switch status {
	case base.ProjectStatusActive:
		targetAppStatus = base.AppStatusActive
	case base.ProjectStatusDisabled:
		targetAppStatus = base.AppStatusDisabled
	case base.ProjectStatusDeleting, base.ProjectStatusMissing:
		// Do nothing
	}

	for _, app := range projectEnv.Apps {
		if targetAppStatus == "" {
			continue
		}
		app.Project = projectEnv.Project
		app.ProjectEnv = projectEnv
		// Run app update in a separate transaction to reduce lock time
		var changed []*entity.App
		err = s.ExecuteEnvInTx(ctx, projectEnv, true, func(db database.Tx) (err error) {
			changed, err = s.appService.SetAppStatus(ctx, db, app, targetAppStatus, recursive)
			return hperrors.Wrap(err)
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		changedApps = append(changedApps, changed...)
	}

	projectEnv.Status = status
	projectEnv.UpdatedAt = timeutil.NowUTC()
	projectEnv.UpdateVer++

	err = s.projectEnvRepo.Update(ctx, db, projectEnv,
		bunex.UpdateColumns("status", "updated_at", "update_ver"))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return changedApps, nil
}
