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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// labelProjectID names the project an env network was made for. The name alone
// does not: a project made again under the key of a deleted one has the same.
const labelProjectID = "hivepaas.project.id"

// A deleted project's env network lingers until the last of its containers has
// stopped - ten seconds of grace unless an app asks for more - and while it does,
// a service cannot use it ("network ... not found") nor a network take its name
// ("already exists"). networkLingerMax bounds the wait for it to go.
var (
	networkLingerPoll = 500 * time.Millisecond
	networkLingerMax  = 30 * time.Second
)

func (s *service) GetProjectNetworkName(project *entity.Project, env string) string {
	return networkservice.ProjectNetworkName(project, env)
}

func (s *service) GetOrCreateProjectNetwork(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
	env string,
) (*entity.Setting, *network.Inspect, error) {
	netName := s.GetProjectNetworkName(project, env)
	setting, err := s.settingRepo.GetByName(ctx, db, project.GetObjectScope(),
		base.SettingTypeClusterNetwork, netName, true,
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil, hperrors.Wrap(err)
	}
	inspect, err := s.dockerManager.NetworkInspect(ctx, netName)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil, hperrors.Wrap(err)
	}
	// One the project neither records nor labeled is what a deleted project of
	// the same key left: waited out, then made anew.
	if inspect != nil && setting == nil && inspect.Network.Labels[labelProjectID] != project.ID {
		if inspect, err = s.awaitNetworkGone(ctx, netName); err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
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
					labelProjectID:             project.ID,
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
	setting.RefID = inspect.Network.ID
	netEntity := &entity.ClusterNetwork{}
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

// awaitNetworkGone waits for a network a deleted project left to go, and
// answers nil once it has. One still there after networkLingerMax is answered:
// its removal failed rather than lags, and it can be used.
func (s *service) awaitNetworkGone(ctx context.Context, name string) (*client.NetworkInspectResult, error) {
	deadline := time.Now().Add(networkLingerMax)
	for {
		timer := time.NewTimer(networkLingerPoll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, hperrors.Wrap(ctx.Err())
		case <-timer.C:
		}
		inspect, err := s.dockerManager.NetworkInspect(ctx, name)
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if time.Now().After(deadline) {
			return inspect, nil
		}
	}
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
		netIDs = append(netIDs, setting.RefID)
	}

	netList, err := s.dockerManager.NetworkListByIDs(ctx, netIDs)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	networks = make(map[string]*network.Summary, len(netList.Items))
	for i := range netList.Items {
		networks[netList.Items[i].ID] = &netList.Items[i]
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
		net := networks[setting.RefID]
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
