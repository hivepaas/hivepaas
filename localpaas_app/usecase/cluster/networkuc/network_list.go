package networkuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *NetworkUC) ListNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.ListNetworkReq,
) (*networkdto.ListNetworkResp, error) {
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
			return label == "" || label == req.ProjectID
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
