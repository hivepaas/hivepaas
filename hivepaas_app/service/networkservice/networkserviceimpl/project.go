package networkserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) GetProjectNetworkName(project *entity.Project, env string) string {
	if env == "" {
		return project.Key + "_local_net"
	}
	return project.Key + "_" + projecthelper.CalcProjectEnvKey(env) + "_net"
}

func (s *service) GetOrCreateProjectNetwork(
	ctx context.Context,
	project *entity.Project,
	env string,
) (*network.Inspect, error) {
	netName := s.GetProjectNetworkName(project, env)
	inspect, err := s.dockerManager.NetworkInspect(ctx, netName)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New(err)
	}

	if inspect == nil { // not found, create one
		_, err = s.dockerManager.NetworkCreate(ctx, netName,
			func(opts *client.NetworkCreateOptions) {
				opts.Driver = docker.NetworkDriverOverlay
				opts.Scope = docker.NetworkScopeSwarm
				opts.Attachable = true
				opts.Labels = map[string]string{
					docker.StackLabelNamespace: project.Key,
				}
			})
		if err != nil {
			return nil, apperrors.New(err)
		}
		// Inspect again
		inspect, err = s.dockerManager.NetworkInspect(ctx, netName)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	return &inspect.Network, nil
}

func (s *service) ListProjectNetworks(
	ctx context.Context,
	project *entity.Project,
) ([]network.Summary, error) {
	resp, err := s.dockerManager.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "label", docker.StackLabelNamespace+"="+project.Key)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp.Items, nil
}

func (s *service) RemoveProjectNetwork(
	ctx context.Context,
	project *entity.Project,
	env string,
) error {
	_, err := s.dockerManager.NetworkRemove(ctx, s.GetProjectNetworkName(project, env))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.New(err)
	}
	return nil
}

func (s *service) RemoveAllProjectNetworks(
	ctx context.Context,
	project *entity.Project,
) error {
	networks, err := s.ListProjectNetworks(ctx, project)
	if err != nil {
		return apperrors.New(err)
	}
	for i := range networks {
		net := &networks[i]
		_, e := s.dockerManager.NetworkRemove(ctx, net.ID)
		err = errors.Join(err, e)
	}
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
