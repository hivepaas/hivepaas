package projectserviceimpl

import (
	"context"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) DeleteProjectEnv(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv) error {
	// Remove all apps
	var wg sync.WaitGroup
	for _, app := range projectEnv.Apps {
		if app.IsChildApp() {
			continue
		}
		app.ProjectEnv = projectEnv
		app.Project = projectEnv.Project
		wg.Go(func() {
			_ = s.appService.ExecuteInTx(ctx, app, true, func(db database.Tx) error {
				if err := s.appService.DeleteApp(ctx, db, app); err != nil {
					return apperrors.Wrap(err)
				}
				return nil
			})
			// NOTE: it's hard to rollback, maybe we only show the errors if there is any
		})
	}
	wg.Wait()

	// Remove all project env local networks
	err := s.networkService.RemoveAllProjectEnvNetworks(ctx, db, projectEnv)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Remove all project env local volumes
	err = s.volumeService.RemoveAllProjectEnvVolumes(ctx, db, projectEnv, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete ref resources in DB
	projectEnvIDs := []string{projectEnv.ID}

	// ACL permissions related to the project env
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project env tags
	err = s.tagRepo.DeleteAllByObjects(ctx, db, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project env files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProjectEnv, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeProjectEnv, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProjectEnv, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByProjectEnvs(ctx, db, projectEnvIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	projectEnv.DeletedAt = time.Now()
	projectEnv.UpdateVer++
	err = s.projectEnvRepo.Update(ctx, db, projectEnv, bunex.UpdateColumns("deleted_at", "update_ver"))
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
