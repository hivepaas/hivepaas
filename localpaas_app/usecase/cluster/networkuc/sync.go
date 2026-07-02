package networkuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
)

func (uc *UC) SyncNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	_ *networkdto.SyncNetworkReq,
) (*networkdto.SyncNetworkResp, error) {
	// 1. Scan docker to get list of networks
	res, err := uc.dockerManager.NetworkList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// 2. Get list of existing settings from DB
	dbSettings, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	existingNetIDs := make(map[string]bool, len(dbSettings))
	for _, s := range dbSettings {
		netEntity, err := s.AsClusterNetwork()
		if err != nil {
			return nil, apperrors.New(err)
		}
		existingNetIDs[netEntity.NetworkID] = true
	}

	// 3. For each docker network, if not exists in DB, create new setting
	var newSettings []*entity.Setting
	for i := range res.Items {
		net := &res.Items[i]
		if existingNetIDs[net.ID] {
			continue
		}

		// Insert setting
		netEntity := &entity.ClusterNetwork{
			NetworkID: net.ID,
			Name:      net.Name,
		}

		setting := &entity.Setting{
			ID:      gofn.Must(ulid.NewStringULID()),
			Scope:   base.ObjectScopeGlobal,
			Type:    currentSettingType,
			Kind:    net.Driver,
			Status:  base.SettingStatusActive,
			Name:    net.Name,
			Version: currentSettingVersion,
		}
		if err := setting.SetData(netEntity); err != nil {
			return nil, apperrors.New(err)
		}
		newSettings = append(newSettings, setting)
	}

	if err := uc.SettingRepo.InsertMulti(ctx, uc.DB, newSettings); err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.SyncNetworkResp{}, nil
}
