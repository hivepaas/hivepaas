package projectserviceimpl

import (
	"context"
	"errors"
	"sync"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) DeleteProject(ctx context.Context, db database.IDB, project *entity.Project) error {
	// Remove all apps
	var wg sync.WaitGroup
	for _, app := range project.Apps {
		wg.Go(func() {
			_ = s.appService.DeleteApp(ctx, db, app)
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
	err = s.projectTagRepo.DeleteAllByProjects(ctx, db, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeProject, projectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.SubjectTypeProject, projectIDs)
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
	err = s.networkService.RemoveAllProjectNetworks(ctx, project)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
