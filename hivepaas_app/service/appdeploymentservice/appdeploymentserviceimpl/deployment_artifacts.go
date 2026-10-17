package appdeploymentserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// prepareArtifacts has an app's artifacts on the nodes its tasks may start on,
// before its service is updated, and returns the mounts that bind them; none for
// an app without artifacts, so that a stale one is taken off.
//
// It is a download, so it runs before the update rather than inside it, where a
// retry would run it again. Unlike prepareDockerAPI, a failure fails the
// deployment: a task whose bind source is missing does not start.
//
// The nodes are the service's placement as it stands. A managed app's placement
// is its data volume's pin, which a deployment does not move.
func (s *service) prepareArtifacts(
	ctx context.Context,
	db database.IDB,
	data *appDeploymentData,
) ([]mount.Mount, error) {
	artifacts, err := s.artifactService.ArtifactsOf(ctx, db, data.App.ID)
	if err != nil || artifacts == nil {
		return nil, hperrors.Wrap(err)
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	mounts, err := s.artifactService.MountsFor(ctx, artifacts, &inspect.Service.Spec)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return mounts, nil
}
