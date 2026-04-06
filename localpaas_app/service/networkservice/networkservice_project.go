package networkservice

import (
	"context"

	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/services/docker"
)

func (s *networkService) CreateProjectNetwork(
	ctx context.Context,
	project *entity.Project,
) (*network.CreateResponse, error) {
	// Create a default network for the project apps
	net, err := s.dockerManager.NetworkCreate(ctx, project.GetDefaultNetworkName(), func(opts *network.CreateOptions) {
		opts.Driver = docker.NetworkDriverOverlay
		opts.Scope = docker.NetworkScopeSwarm
		opts.Attachable = true
		opts.Labels = map[string]string{
			docker.StackLabelNamespace: project.Key,
		}
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return net, nil
}

func (s *networkService) RemoveProjectNetwork(
	ctx context.Context,
	project *entity.Project,
) error {
	err := s.dockerManager.NetworkRemove(ctx, project.GetDefaultNetworkName())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
