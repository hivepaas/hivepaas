package networkuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) ListNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.ListNetworkReq,
) (_ *networkdto.ListNetworkResp, err error) {
	var project *entity.Project
	if req.ProjectID != "" {
		project, err = uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	networks, err := uc.dockerManager.NetworkList(ctx, func(opts *network.ListOptions) {
		if opts.Filters.Len() == 0 {
			opts.Filters = filters.NewArgs()
		}
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	filterNetworks := networks
	if req.ProjectID != "" {
		filterNetworks = gofn.FilterPtr(filterNetworks, func(net *network.Summary) bool {
			label := net.Labels[docker.StackLabelNamespace]
			return label == "" || label == project.Key
		})
	}
	if req.Search != "" {
		keyword := strings.ToLower(req.Search)
		filterNetworks = gofn.FilterPtr(filterNetworks, func(net *network.Summary) bool {
			return strings.Contains(strings.ToLower(net.Name), keyword)
		})
	}
	if len(auth.AllowObjectIDs) > 0 {
		filterNetworks = gofn.FilterPtr(filterNetworks, func(net *network.Summary) bool {
			return gofn.Contain(auth.AllowObjectIDs, net.ID) || gofn.Contain(auth.AllowObjectIDs, net.Name)
		})
	}

	return &networkdto.ListNetworkResp{
		Meta: &basedto.ListMeta{Page: &basedto.PagingMeta{
			Offset: 0,
			Limit:  req.Paging.Limit,
			Total:  len(filterNetworks),
		}},
		Data: networkdto.TransformNetworks(filterNetworks),
	}, nil
}
