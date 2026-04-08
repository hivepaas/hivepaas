package nodeuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) ListNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.ListNodeReq,
) (*nodedto.ListNodeResp, error) {
	nodes, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	filterNodes := nodes
	if len(req.Status) > 0 {
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return gofn.Contain(req.Status, base.NodeStatus(node.Status.State))
		})
	}
	if len(req.Role) > 0 {
		filterNodes = gofn.FilterPtr(filterNodes, func(node *swarm.Node) bool {
			return gofn.Contain(req.Role, base.NodeRole(node.Spec.Role))
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
