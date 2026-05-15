package projectserviceimpl

import (
	"context"
	"errors"
	"sync"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

func (s *service) DeleteProject(ctx context.Context, project *entity.Project) error {
	// Remove all apps
	var wg sync.WaitGroup
	for _, app := range project.Apps {
		wg.Go(func() {
			_ = s.appService.DeleteApp(ctx, app)
			// NOTE: it's hard to rollback, maybe we only show the errors if there is any
		})
	}
	wg.Wait()

	// Remove all project local networks
	err := s.networkService.RemoveAllProjectNetworks(ctx, project)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
