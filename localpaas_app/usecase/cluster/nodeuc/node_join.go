package nodeuc

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/localpaas/localpaas/services/ssh"
)

const (
	executionTimeout = time.Second * 15
)

func (uc *UC) JoinNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.JoinNodeReq,
) (*nodedto.JoinNodeResp, error) {
	data := &joinNodeData{}
	err := uc.loadJoinNodeData(ctx, uc.db, req, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	cmdCtx, cancelFunc := context.WithTimeout(ctx, executionTimeout)
	defer cancelFunc()

	command := fmt.Sprintf("docker swarm leave --force && docker swarm join --token %s %s",
		data.JoinToken, data.PreferManagerAddr)
	output, err := ssh.Execute(cmdCtx, &ssh.CommandInput{
		Host:       req.Host,
		Port:       req.Port,
		User:       req.User,
		PrivateKey: data.SSHKey.PrivateKey.MustGetPlain(),
		Passphrase: data.SSHKey.Passphrase.MustGetPlain(),
		Command:    command,
	})
	if err != nil {
		return nil, apperrors.New(apperrors.ErrInfraActionFailed).WithParam("Error", err.Error())
	}

	return &nodedto.JoinNodeResp{
		Data: &nodedto.JoinNodeDataResp{
			CommandOutput: output,
		},
	}, nil
}

type joinNodeData struct {
	SSHKey            *entity.SSHKey
	JoinToken         string
	PreferManagerAddr string
}

func (uc *UC) loadJoinNodeData(
	ctx context.Context,
	db database.IDB,
	req *nodedto.JoinNodeReq,
	data *joinNodeData,
) error {
	sshKeySetting, err := uc.settingRepo.GetByID(ctx, db, base.NewSettingScopeGlobal(), base.SettingTypeSSHKey,
		req.SSHKey.ID, true)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.SSHKey = sshKeySetting.MustAsSSHKey()

	// Find join token from the cluster
	theSwarm, err := uc.dockerManager.SwarmInspect(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	joinToken := gofn.If(req.JoinAsManager, theSwarm.JoinTokens.Manager, theSwarm.JoinTokens.Worker)
	if joinToken == "" {
		return apperrors.New(apperrors.ErrInfraInternal).
			WithNTParam("Error", "join token is not found")
	}
	data.JoinToken = joinToken

	// List all manager nodes to get the addr to join new node
	managerNodes, err := uc.dockerManager.NodeList(ctx, func(opts *swarm.NodeListOptions) {
		opts.Filters = filters.NewArgs(filters.Arg("role", "manager"))
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	var leaderAddr, managerAddr string
	for _, node := range managerNodes {
		mgrStatus := node.ManagerStatus
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
