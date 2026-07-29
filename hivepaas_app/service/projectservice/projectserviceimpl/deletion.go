package projectserviceimpl

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) DeleteProject(ctx context.Context, db database.IDB, project *entity.Project) error {
	// Remove all envs
	var wg sync.WaitGroup
	for _, env := range project.ProjectEnvs {
		env.Project = project
		wg.Go(func() {
			_ = s.ExecuteEnvInTx(ctx, env, true, func(db database.Tx) error {
				if err := s.DeleteProjectEnv(ctx, db, env); err != nil {
					return apperrors.Wrap(err)
				}
				return nil
			})
			// NOTE: it's hard to rollback, maybe we only show the errors if there is any
		})
	}
	wg.Wait()

	// Delete ref resources in DB
	projectIDs := []string{project.ID}

	// ACL permissions having the project ID as subject ID
	err := s.permissionManager.RemoveACLPermissionsBySubjects(ctx, db, base.SubjectTypeProject, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project tags
	err = s.tagRepo.DeleteAllByObjects(ctx, db, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeProject, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByProjects(ctx, db, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project photo
	if project.PhotoID != "" {
		err = s.binObjectRepo.DeleteByIDs(ctx, db, []string{project.PhotoID})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Remove all project local networks
	err = s.networkService.RemoveAllProjectNetworks(ctx, db, project)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	project.DeletedAt = time.Now()
	project.UpdateVer++
	err = s.projectRepo.Update(ctx, db, project, bunex.UpdateColumns("deleted_at", "update_ver"))
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) DeleteProjectEnv(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv) error {
	// Remove all apps
	var wg sync.WaitGroup
	for _, app := range projectEnv.Apps {
		if app.ParentID != "" {
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

	// Delete ref resources in DB
	projectEnvIDs := []string{projectEnv.ID}

	// ACL permissions having the project env ID as subject ID
	err := s.permissionManager.RemoveACLPermissionsBySubjects(ctx, db, base.SubjectTypeProjectEnv, projectEnvIDs)
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

	// Remove all project env local networks
	err = s.networkService.RemoveAllProjectEnvNetworks(ctx, db, projectEnv)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
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
