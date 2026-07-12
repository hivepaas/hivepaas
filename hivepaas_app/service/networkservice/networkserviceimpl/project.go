package networkserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) GetProjectNetworkName(project *entity.Project, env string) string {
	if env == "" {
		return project.Key + "_local_net"
	}
	return project.Key + "_" + projecthelper.CalcProjectEnvKey(env) + "_net"
}

func (s *service) GetOrCreateProjectNetwork(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
	env string,
) (*entity.Setting, *network.Inspect, error) {
	netName := s.GetProjectNetworkName(project, env)
	inspect, err := s.dockerManager.NetworkInspect(ctx, netName)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil, apperrors.New(err)
	}

	if inspect == nil { // not found, create one
		_, err = s.dockerManager.NetworkCreate(ctx, netName,
			func(opts *client.NetworkCreateOptions) {
				opts.Driver = docker.NetworkDriverOverlay
				opts.Scope = docker.NetworkScopeSwarm
				opts.Attachable = true
				opts.Labels = map[string]string{
					docker.StackLabelNamespace: project.Key,
				}
			})
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
		// Inspect again
		inspect, err = s.dockerManager.NetworkInspect(ctx, netName)
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
	}

	setting, err := s.settingRepo.GetByName(ctx, db, project.GetObjectScope(),
		base.SettingTypeClusterNetwork, netName, true,
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil, apperrors.New(err)
	}
	hasChange := false
	if setting == nil {
		hasChange = true
		timeNow := time.Now()
		setting = &entity.Setting{
			ID:        dockerhelper.WrapNetworkID(inspect.Network.ID),
			Scope:     base.ObjectScopeProject,
			ObjectID:  project.ID,
			Type:      base.SettingTypeClusterNetwork,
			Status:    base.SettingStatusActive,
			Default:   true,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	if setting.Kind != inspect.Network.Driver {
		hasChange = true
		setting.Kind = inspect.Network.Driver
	}
	if setting.Name != inspect.Network.Name {
		hasChange = true
		setting.Name = inspect.Network.Name
	}
	if err = setting.SetData(&entity.ClusterNetwork{}); err != nil {
		return nil, nil, apperrors.New(err)
	}

	if hasChange {
		err = s.settingRepo.Upsert(ctx, db, setting,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
	}

	return setting, &inspect.Network, nil
}

func (s *service) ListProjectNetworks(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) (settings []*entity.Setting, networks map[string]*network.Summary, err error) {
	settings, _, err = s.settingRepo.List(ctx, db, project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNetwork),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	if len(settings) == 0 {
		return nil, nil, nil
	}

	netIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		netIDs = append(netIDs, dockerhelper.ParseID(setting.ID))
	}

	netList, err := s.dockerManager.NetworkListByIDs(ctx, netIDs)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	networks = make(map[string]*network.Summary, len(settings))
	for _, netID := range netIDs {
		net, found := gofn.FindPtr(netList.Items, func(net *network.Summary) bool {
			return net.ID == netID
		})
		if found {
			networks[net.ID] = &net
		}
	}

	return settings, networks, nil
}

func (s *service) RemoveAllProjectNetworks(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) error {
	settings, networks, err := s.ListProjectNetworks(ctx, db, project)
	if err != nil {
		return apperrors.New(err)
	}

	for _, setting := range settings {
		if setting.ObjectID != project.ID { // imported/inherited network, skip it
			continue
		}
		net := networks[dockerhelper.ParseID(setting.ID)]
		if net == nil {
			continue
		}
		_, e := s.dockerManager.NetworkRemove(ctx, net.ID)
		err = errors.Join(err, e)
	}
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
