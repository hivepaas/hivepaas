package nodeuc

import (
	"context"
	"fmt"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) GetNodeJoinCommand(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeJoinCommandReq,
) (*nodedto.GetNodeJoinCommandResp, error) {
	// The join token is a secret, and this endpoint hands it over in the clear.
	// It is not stored as a setting - it comes from the Swarm itself - which is
	// the only reason it did not already pass through this gate: the same two
	// checks and the same record apply, because anybody holding the token can put
	// a machine into the cluster, and the manager token puts it in with control
	// over everything.
	//
	// Gated before the token is read, so a refusal never reaches Docker.
	role := "worker"
	if req.JoinAsManager {
		role = "manager"
	}
	detail := auditdetail.New().Set("tokenRole", role).String()
	err := uc.AuthorizeSecretReveal(ctx, uc.DB, auth, &settings.RevealSubject{
		Scope:      base.ObjectScopeGlobal,
		Source:     base.AuditLogSourceAPIGet,
		SecretType: base.SecretTypeSwarmJoinToken,
		ResType:    base.ResourceTypeCluster,
		Detail:     detail,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	data := &joinNodeCommandData{}
	err = uc.loadGetNodeJoinCommandData(ctx, req, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}
	theSwarm := &inspect.Swarm

	joinToken := gofn.If(req.JoinAsManager, theSwarm.JoinTokens.Manager, theSwarm.JoinTokens.Worker)
	if joinToken == "" {
		return hperrors.Wrap(hperrors.ErrInfraInternal).
			WithParam("Error", "join token is not found")
	}
	data.JoinToken = joinToken

	// List all manager nodes to get the addr to join new node
	listResp, err := uc.dockerManager.NodeManagerList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(hperrors.ErrInfraInternal).
			WithParam("Error", "active manager node not found")
	}

	return nil
}
