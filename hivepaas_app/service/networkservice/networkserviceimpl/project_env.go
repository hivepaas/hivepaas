package networkserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
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
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil, hperrors.Wrap(err)
	}

	if inspect == nil { // not found, create one
		_, err = s.dockerManager.NetworkCreate(ctx, netName,
			func(opts *client.NetworkCreateOptions) {
				opts.Driver = docker.NetworkDriverOverlay
				opts.Scope = docker.NetworkScopeSwarm
				opts.Attachable = true
				opts.Options = map[string]string{
					docker.NetworkOptionDriverMTU: docker.DefaultOverlayNetworkMTU,
				}
				opts.Labels = map[string]string{
					docker.StackLabelNamespace: project.Key,
				}
			})
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		// Inspect again
		inspect, err = s.dockerManager.NetworkInspect(ctx, netName)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
	}

	setting, err := s.settingRepo.GetByName(ctx, db, project.GetObjectScope(),
		base.SettingTypeClusterNetwork, netName, true,
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil, hperrors.Wrap(err)
	}
	hasChange := false
	if setting == nil {
		hasChange = true
		timeNow := time.Now()
		setting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeProject,
			ObjectID:    project.ID,
			Type:        base.SettingTypeClusterNetwork,
			Status:      base.SettingStatusActive,
			Inheritable: true,
			Default:     true,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
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
	netEntity := &entity.ClusterNetwork{
		RefID: inspect.Network.ID,
	}
	if err = setting.SetData(netEntity); err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	if hasChange {
		err = s.settingRepo.Upsert(ctx, db, setting,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
	}

	return setting, &inspect.Network, nil
}

func (s *service) ListProjectEnvNetworks(
	ctx context.Context,
	db database.IDB,
	projectEnv *entity.ProjectEnv,
) (settings []*entity.Setting, networks map[string]*network.Summary, err error) {
	settings, _, err = s.settingRepo.List(ctx, db, projectEnv.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNetwork),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil, nil
	}

	netIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		netIDs = append(netIDs, setting.MustAsClusterNetwork().RefID)
	}

	netList, err := s.dockerManager.NetworkListByIDs(ctx, netIDs)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
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

func (s *service) RemoveAllProjectEnvNetworks(
	ctx context.Context,
	db database.IDB,
	projectEnv *entity.ProjectEnv,
) error {
	settings, networks, err := s.ListProjectEnvNetworks(ctx, db, projectEnv)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, setting := range settings {
		if setting.ObjectID != projectEnv.ID { // imported/inherited network, skip it
			continue
		}
		net := networks[setting.MustAsClusterNetwork().RefID]
		if net == nil {
			continue
		}
		_, e := s.dockerManager.NetworkRemove(ctx, net.ID)
		if e != nil && !errors.Is(e, hperrors.ErrNotFound) {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
