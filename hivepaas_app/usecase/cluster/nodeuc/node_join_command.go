package nodeuc

import (
	"context"
	"fmt"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) GetNodeJoinCommand(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeJoinCommandReq,
) (*nodedto.GetNodeJoinCommandResp, error) {
	data := &joinNodeCommandData{}
	err := uc.loadGetNodeJoinCommandData(ctx, req, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	command := fmt.Sprintf("docker swarm join --token %s %s", data.JoinToken, data.PreferManagerAddr)
	return &nodedto.GetNodeJoinCommandResp{
		Data: &nodedto.GetNodeJoinCommandDataResp{
			Command: command,
		},
	}, nil
}

type joinNodeCommandData struct {
	JoinToken         string
	PreferManagerAddr string
}

func (uc *UC) loadGetNodeJoinCommandData(
	ctx context.Context,
	req *nodedto.GetNodeJoinCommandReq,
	data *joinNodeCommandData,
) error {
	// Find join token from the cluster
	inspect, err := uc.dockerManager.SwarmInspect(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	theSwarm := &inspect.Swarm

	joinToken := gofn.If(req.JoinAsManager, theSwarm.JoinTokens.Manager, theSwarm.JoinTokens.Worker)
	if joinToken == "" {
		return apperrors.New(apperrors.ErrInfraInternal).
			WithNTParam("Error", "join token is not found")
	}
	data.JoinToken = joinToken

	// List all manager nodes to get the addr to join new node
	listResp, err := uc.dockerManager.NodeManagerList(ctx)
	if err != nil {
		return apperrors.New(err)
	}

	var leaderAddr, managerAddr string
	for i := range listResp.Items {
		mgrStatus := listResp.Items[i].ManagerStatus
		if mgrStatus.Reachability == swarm.ReachabilityReachable {
			managerAddr = mgrStatus.Addr
			if mgrStatus.Leader {
				leaderAddr = mgrStatus.Addr
			}
		}
	}
	data.PreferManagerAddr = gofn.Coalesce(leaderAddr, managerAddr)
	if data.PreferManagerAddr == "" {
		return apperrors.New(apperrors.ErrInfraInternal).
			WithNTParam("Error", "active manager node not found")
	}

	return nil
}
