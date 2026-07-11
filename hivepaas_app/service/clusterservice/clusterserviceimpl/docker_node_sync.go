package clusterserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
)

func (s *service) SyncNodes(
	ctx context.Context,
	db database.IDB,
) ([]swarm.Node, error) {
	// 1. Scan docker to get list of nodes
	nodeList, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	currentSettingType := base.SettingTypeClusterNode
	currentSettingVersion := entity.CurrentClusterNodeVersion

	// 2. Get list of existing settings from DB
	dbSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	existingNodes := make(map[string]*entity.Setting, len(dbSettings))
	for _, s := range dbSettings {
		existingNodes[dockerhelper.ParseID(s.ID)] = s
	}

	// 3. For each docker node, if not exists in DB, create new setting
	var updatingSettings []*entity.Setting
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		setting := existingNodes[node.ID]

		if setting == nil {
			setting = &entity.Setting{
				ID:      dockerhelper.WrapNodeID(node.ID),
				Scope:   base.ObjectScopeGlobal,
				Type:    currentSettingType,
				Kind:    string(node.Spec.Role),
				Status:  base.SettingStatusActive,
				Name:    node.Spec.Name,
				Version: currentSettingVersion,
			}
			nodeEntity := &entity.ClusterNode{}
			if err := setting.SetData(nodeEntity); err != nil {
				return nil, apperrors.New(err)
			}
			updatingSettings = append(updatingSettings, setting)
			continue
		}

		delete(existingNodes, node.ID)
		hasChanged := false
		if setting.Kind != string(node.Spec.Role) {
			setting.Kind = string(node.Spec.Role)
			hasChanged = true
		}
		if setting.Name != node.Spec.Name {
			setting.Name = node.Spec.Name
			hasChanged = true
		}
		if hasChanged {
			updatingSettings = append(updatingSettings, setting)
		}
	}

	// 4. All node settings that exist in DB but docker swarm need to remove
	timeNow := time.Now()
	for _, s := range existingNodes {
		s.DeletedAt = timeNow
		updatingSettings = append(updatingSettings, s)
	}

	// 5. Upsert the settings
	err = s.settingRepo.UpsertMulti(ctx, db, updatingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return nodeList.Items, nil
}
