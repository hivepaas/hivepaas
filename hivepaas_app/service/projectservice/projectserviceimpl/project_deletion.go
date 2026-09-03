package projectserviceimpl

import (
	"context"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

func (s *service) DeleteProject(ctx context.Context, db database.IDB, project *entity.Project) error {
	// Remove all envs
	var wg sync.WaitGroup
	for _, env := range project.ProjectEnvs {
		env.Project = project
		wg.Go(func() {
			defer safego.Recover("projectservice.deleteProjectEnv")
			_ = s.ExecuteEnvInTx(ctx, env, true, func(db database.Tx) error {
				if err := s.DeleteProjectEnv(ctx, db, env); err != nil {
					return hperrors.Wrap(err)
				}
				return nil
			})
			// NOTE: it's hard to rollback, maybe we only show the errors if there is any
		})
	}
	wg.Wait()

	// Remove all project local networks
	err := s.networkService.RemoveAllProjectNetworks(ctx, db, project)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Remove all project local volumes
	err = s.volumeService.RemoveAllProjectVolumes(ctx, db, project, false)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Delete ref resources in DB
	projectIDs := []string{project.ID}

	// ACL permissions related to the project
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project tags
	err = s.tagRepo.DeleteAllByObjects(ctx, db, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeProject, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByProjects(ctx, db, projectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project photo
	if project.Photo != "" {
		err = s.binObjectRepo.DeleteByIDs(ctx, db, []string{project.Photo})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	project.DeletedAt = time.Now()
	project.UpdateVer++
	err = s.projectRepo.Update(ctx, db, project, bunex.UpdateColumns("deleted_at", "update_ver"))
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
