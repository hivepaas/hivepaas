package dockerapiserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) AccessOf(ctx context.Context, db database.IDB, appID string) (
	*entity.AppDockerAPISettings, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDockerAPI),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}
	access, err := settings[0].AsAppDockerAPISettings()
	return access, hperrors.Wrap(err)
}

func (s *service) AppNetworkID(ctx context.Context, appID string) (string, error) {
	inspect, err := s.dockerManager.NetworkInspect(ctx, dockerapiservice.NetworkName(appID))
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return "", nil
		}
		return "", hperrors.Wrap(err)
	}
	return inspect.Network.ID, nil
}

// EnsureNetwork makes the app's own network an attachable overlay: the app's
// task is a swarm service and joins it that way, and its children are plain
// containers on whichever node the task runs.
func (s *service) EnsureNetwork(ctx context.Context, appID string) (string, error) {
	id, err := s.AppNetworkID(ctx, appID)
	if err != nil || id != "" {
		return id, err
	}
	resp, err := s.dockerManager.NetworkCreate(ctx, dockerapiservice.NetworkName(appID),
		func(opts *client.NetworkCreateOptions) {
			opts.Driver = docker.NetworkDriverOverlay
			opts.Scope = docker.NetworkScopeSwarm
			opts.Attachable = true
			opts.Options = map[string]string{docker.NetworkOptionDriverMTU: docker.DefaultOverlayNetworkMTU}
			opts.Labels = map[string]string{dockerapiservice.NetworkLabel: appID}
		})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return resp.ID, nil
}

func (s *service) ApplyToService(ctx context.Context, db database.IDB, appID string,
	spec *swarm.ServiceSpec) error {
	access, err := s.AccessOf(ctx, db, appID)
	if err != nil {
		return err
	}
	if access == nil {
		return s.DetachFromService(ctx, appID, spec)
	}
	networkID, err := s.EnsureNetwork(ctx, appID)
	if err != nil {
		return err
	}
	dockerapiservice.Attach(spec, appID, networkID)
	return nil
}

func (s *service) DetachFromService(ctx context.Context, appID string, spec *swarm.ServiceSpec) error {
	networkID, err := s.AppNetworkID(ctx, appID)
	if err != nil {
		return err
	}
	dockerapiservice.Detach(spec, networkID)
	return nil
}

// RemoveApp does what it can. It runs once the app's service is removed, and
// the agents' sweep and a later delete of the network take care of what a node
// that is down, or a task still shutting down, keeps for now.
func (s *service) RemoveApp(ctx context.Context, appID string) error {
	errs := []error{s.RemoveAppObjects(ctx, appID)}
	_, err := s.dockerManager.NetworkRemove(ctx, dockerapiservice.NetworkName(appID))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		errs = append(errs, hperrors.Wrap(err))
	}
	return errors.Join(errs...)
}
