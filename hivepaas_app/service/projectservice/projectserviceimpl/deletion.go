package projectserviceimpl

import (
	"context"
	"errors"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) DeleteProject(ctx context.Context, db database.IDB, project *entity.Project) error {
	// Remove all apps
	var wg sync.WaitGroup
	for _, app := range project.Apps {
		wg.Go(func() {
			_ = s.appService.ExecuteInTx(ctx, app, true, func(db database.Tx) error {
				if err := s.appService.DeleteApp(ctx, db, app); err != nil {
					return apperrors.New(err)
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
		return apperrors.New(err)
	}

	// Project tags
	err = s.projectTagRepo.DeleteAllByProjects(ctx, db, projectIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Project files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeProject, projectIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByProjects(ctx, db, projectIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Project photo
	if project.PhotoID != "" {
		err = s.binObjectRepo.DeleteByIDs(ctx, db, []string{project.PhotoID})
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Remove all project local networks
	err = s.networkService.RemoveAllProjectNetworks(ctx, db, project)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}

	return nil
}
