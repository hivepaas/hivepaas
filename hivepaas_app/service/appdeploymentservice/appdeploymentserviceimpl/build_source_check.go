package appdeploymentserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/gittool"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
)

func (s *service) CheckBuildSource(
	ctx context.Context,
	req *appdeploymentservice.CheckBuildSourceReq,
) error {
	// When the cluster has multiple nodes, the result image must be pushed to a registry
	// that can be accessed by all the nodes in the cluster.
	isMultiNode, err := s.clusterService.IsMultiNode(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if isMultiNode && req.PushToRegistry.ID == "" {
		return hperrors.Wrap(hperrors.ErrMultiNodeClusterRequireRegistryForImages)
	}

	repoSource := req.RepoSource
	if repoSource == nil {
		return nil
	}
	// Validate existence of repo and ref
	switch repoSource.RepoType { //nolint:gocritic
	case base.RepoTypeGit:
		// TODO: do not check commit hash for now, that's so slow
		err := gittool.ValidateWithGitCli(ctx, &gittool.ValidationOptions{
			URL:           repoSource.RepoURL,
			Credentials:   req.RefObjects.RefSettings[repoSource.Credentials.ID],
			ReferenceName: githelper.ReferenceName(repoSource.RepoRef),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
