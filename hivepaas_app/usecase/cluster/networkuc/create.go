package networkuc

import (
	"context"
	"errors"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) CreateNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.CreateNetworkReq,
) (*networkdto.CreateNetworkResp, error) {
	req.Type = currentSettingType
	netEntity := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingRefIDs: netEntity.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			// If network is for a project, need to apply some restrictions
			if req.Scope.IsProjectScope() {
				req.Driver = docker.NetworkDriverOverlay
				req.Attachable = false
				req.Name = data.ScopeProject.Key + "_" + req.Name
				netEntity.Name = req.Name
			}
			createResp, err := uc.createNetworkInDocker(ctx, req.CreateNetworkBaseReq)
			if err != nil {
				return apperrors.New(err)
			}
			netEntity.NetworkID = createResp.ID

			pData.Setting.Name = req.Name
			pData.Setting.Kind = req.Driver
			if err := pData.Setting.SetData(netEntity); err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.CreateNetworkResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) createNetworkInDocker(
	ctx context.Context,
	req *networkdto.CreateNetworkBaseReq,
) (*client.NetworkCreateResult, error) {
	_, err := uc.dockerManager.NetworkInspect(ctx, req.Name)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New(err)
	}
	if err == nil {
		return nil, apperrors.NewAlreadyExist("Cluster network")
	}

	createResp, err := uc.dockerManager.NetworkCreate(ctx, req.Name, func(opts *client.NetworkCreateOptions) {
		opts.Driver = req.Driver
		opts.Scope = docker.NetworkScopeSwarm
		opts.EnableIPv4 = &req.EnableIPv4
		opts.EnableIPv6 = &req.EnableIPv6
		opts.Internal = req.Internal
		opts.Attachable = req.Attachable
		opts.Options = req.Options
		opts.Labels = req.Labels
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return createResp, nil
}
