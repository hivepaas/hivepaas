package nodeuc

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) ListNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.ListNodeReq,
) (*nodedto.ListNodeResp, error) {
	listResp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	filterNodes := listResp.Items
	if len(req.Status) > 0 {
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return gofn.Contain(req.Status, docker.NodeStatus(node.Status.State))
		})
	}
	if len(req.Role) > 0 {
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return gofn.Contain(req.Role, docker.NodeRole(node.Spec.Role))
		})
	}
	if req.Search != "" {
		keyword := strings.ToLower(req.Search)
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return strings.Contains(node.Description.Hostname, keyword)
		})
	}
	if len(auth.AllowObjectIDs) > 0 {
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return gofn.Contain(auth.AllowObjectIDs, node.ID)
		})
	}

	return &nodedto.ListNodeResp{
		Meta: &basedto.ListMeta{Page: &basedto.PagingMeta{
			Offset: 0,
			Limit:  req.Paging.Limit,
			Total:  len(filterNodes),
		}},
		Data: nodedto.TransformNodes(filterNodes, false),
	}, nil
}
