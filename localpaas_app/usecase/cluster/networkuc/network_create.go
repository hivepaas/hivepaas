package networkuc

import (
	"context"

	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	namespaceGlobal = "global"
)

func (uc *UC) CreateNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.CreateNetworkReq,
) (*networkdto.CreateNetworkResp, error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		req.Labels[docker.StackLabelNamespace] = project.Key
	} else if !req.AvailInProjects {
		req.Labels[docker.StackLabelNamespace] = namespaceGlobal
	}

	resp, err := uc.dockerManager.NetworkCreate(ctx, req.Name, func(opts *network.CreateOptions) {
		opts.Driver = req.Driver
		opts.Scope = docker.NetworkScopeSwarm
		opts.EnableIPv4 = &req.EnableIPv4
		opts.EnableIPv6 = &req.EnableIPv6
		opts.Internal = req.Internal
		opts.Attachable = req.Attachable
		opts.Ingress = req.Ingress
		opts.Options = req.Options
		opts.Labels = req.Labels
	})

	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &networkdto.CreateNetworkResp{
		Data: &basedto.ObjectIDResp{
			ID: resp.ID,
		},
	}, nil
}
