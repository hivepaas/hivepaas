package networkserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) SyncNetworks(
	ctx context.Context,
	db database.IDB,
) ([]network.Summary, error) {
	// 1. Scan docker to get list of networks
	netList, err := s.dockerManager.NetworkList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	currentSettingType := base.SettingTypeClusterNetwork
	currentSettingVersion := entity.CurrentClusterNetworkVersion

	// 2. Get list of existing settings from DB
	dbSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	existingNets := make(map[string]*entity.Setting, len(dbSettings))
	for _, s := range dbSettings {
		existingNets[s.MustAsClusterNetwork().RefID] = s // map docker-net-id -> *Setting
	}

	// 3. For each docker network, if not exists in DB, create new setting
	var updatingSettings []*entity.Setting
	for i := range netList.Items {
		net := &netList.Items[i]
		setting := existingNets[net.ID]

		if setting == nil {
			setting = &entity.Setting{
				ID:      gofn.Must(ulid.NewStringULID()),
				Scope:   base.ObjectScopeGlobal,
				Type:    currentSettingType,
				Kind:    net.Driver,
				Status:  base.SettingStatusActive,
				Name:    net.Name,
				Version: currentSettingVersion,
			}
			volEntity := &entity.ClusterNetwork{
				RefID: net.ID,
			}
			if net.Name == base.NetworkGlobalRouting {
				setting.AvailInProjects = true
			}
			if err := setting.SetData(volEntity); err != nil {
				return nil, apperrors.Wrap(err)
			}
			updatingSettings = append(updatingSettings, setting)
			continue
		}

		delete(existingNets, net.ID)
		hasChanged := false
		if setting.Kind != net.Driver {
			setting.Kind = net.Driver
			hasChanged = true
		}
		if setting.Name != net.Name {
			setting.Name = net.Name
			hasChanged = true
		}
		if hasChanged {
			updatingSettings = append(updatingSettings, setting)
		}
	}

	// 4. All settings that exist in DB but docker swarm need to remove
	timeNow := time.Now()
	for _, s := range existingNets {
		s.DeletedAt = timeNow
		updatingSettings = append(updatingSettings, s)
	}

	// 5. Upsert the settings
	err = s.settingRepo.UpsertMulti(ctx, db, updatingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return netList.Items, nil
}
