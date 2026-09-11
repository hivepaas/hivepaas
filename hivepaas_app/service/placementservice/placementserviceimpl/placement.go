package placementserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

type placementSettingsData struct {
	*placementservice.ApplyPlacementSettingsReq
	IsMultiNode bool
	HasChanges  bool
}

func (s *service) ApplyPlacementSettings(
	ctx context.Context,
	db database.IDB,
	req *placementservice.ApplyPlacementSettingsReq,
) (resp *placementservice.ApplyPlacementSettingsResp, err error) {
	resp = &placementservice.ApplyPlacementSettingsResp{}
	data := &placementSettingsData{
		ApplyPlacementSettingsReq: req,
	}
	err = s.loadPlacementSettingsData(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	s.applyPlacementSettings(data)

	if data.HasChanges && !data.SkipSavingToDocker && data.Service.ID != "" {
		_, err = s.dockerManager.ServiceUpdate(ctx, data.Service.ID, &data.Service.Version, &data.Service.Spec)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	resp.Service = data.Service
	resp.HasChanges = data.HasChanges
	return resp, nil
}

func (s *service) loadPlacementSettingsData(
	ctx context.Context,
	db database.IDB,
	data *placementSettingsData,
) error {
	if data.Service == nil {
		if data.App == nil || data.App.ServiceID == "" {
			return hperrors.NewArgumentInvalid("App.ServiceID")
		}
		inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.Service = &inspect.Service
	}

	isMultiNode, err := s.clusterService.IsMultiNode(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.IsMultiNode = isMultiNode

	defer func() {
		if data.PlacementSettings == nil {
			data.PlacementSettings = &entity.AppPlacementSettings{}
		}
		if data.BuildSettings == nil {
			data.BuildSettings = &entity.ImageBuildSettings{}
		}
	}()

	if !isMultiNode {
		return nil
	}

	scope := data.App.GetObjectScope()

	// Load placement settings
	if data.PlacementSettings == nil {
		placementSetting, err := s.settingRepo.GetSingle(ctx, db, scope,
			base.SettingTypeAppPlacement, false)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if placementSetting != nil && !placementSetting.IsExpired() {
			data.PlacementSettings = placementSetting.MustAsAppPlacementSettings()
		}
	}

	// Load build settings
	if data.BuildSettings == nil {
		buildSetting, err := s.settingRepo.GetSingle(ctx, db, scope,
			base.SettingTypeImageBuild, false)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if buildSetting != nil && !buildSetting.IsExpired() {
			data.BuildSettings = buildSetting.MustAsImageBuildSettings()
		}
	}

	// Load the pins of the volumes this app mounts
	if data.VolumePins == nil {
		pins, err := s.loadVolumePins(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.VolumePins = pins
	}

	s.warnOnPinDrift(ctx, data)

	return nil
}

// loadVolumePins reads the pin of every volume the service mounts, by whatever
// identity that mount type carries.
//
// A TypeVolume/TypeCluster mount names its volume's RefID directly. A TypeBind
// mount - what a pinned `local` volume with type=none/device=<dir> actually
// becomes on the wire, see storage_settings_update.go's
// useBindMountIfAppropriate - carries no volume identity at all, only the host
// path. So this loads every cluster-volume setting the service could possibly
// mount rather than filtering by ref id, and leaves the matching (including the
// bind-to-device comparison) to VolumePinsForMounts.
//
// The scope passed to List is the app's own - the same one
// storage_settings_update.go validates a mount against - so a volume defined at
// a parent scope (project, project env, or global) is found here exactly when
// the app was allowed to mount it: List already walks up the scope chain for
// settings marked inheritable.
func (s *service) loadVolumePins(
	ctx context.Context,
	db database.IDB,
	data *placementSettingsData,
) ([]placementservice.VolumePin, error) {
	mounts := data.Service.Spec.TaskTemplate.ContainerSpec.Mounts
	if len(mounts) == 0 {
		return nil, nil
	}

	settings, _, err := s.settingRepo.List(ctx, db, data.App.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	pins, err := placementservice.VolumePinsForMounts(mounts, settings)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return pins, nil
}
