package placementserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/mount"
	"github.com/uptrace/bun"

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

	return nil
}

// loadVolumePins reads the pin of every volume the service mounts by name.
//
// Mounts are matched on RefID because that is the volume's docker-side identity,
// which is what a mount spec names. The scope passed to List is the app's own -
// the same one storage_settings_update.go validates a mount against - so a
// volume defined at a parent scope (project, project env, or global) is found
// here exactly when the app was allowed to mount it: List already walks up the
// scope chain for settings marked inheritable.
func (s *service) loadVolumePins(
	ctx context.Context,
	db database.IDB,
	data *placementSettingsData,
) ([]placementservice.VolumePin, error) {
	var refIDs []string
	for _, mnt := range data.Service.Spec.TaskTemplate.ContainerSpec.Mounts {
		if mnt.Type == mount.TypeVolume || mnt.Type == mount.TypeCluster {
			refIDs = append(refIDs, mnt.Source)
		}
	}
	if len(refIDs) == 0 {
		return nil, nil
	}

	settings, _, err := s.settingRepo.List(ctx, db, data.App.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
		bunex.SelectWhere("setting.ref_id IN (?)", bun.List(refIDs)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	pins := make([]placementservice.VolumePin, 0, len(settings))
	for _, setting := range settings {
		vol, err := setting.AsClusterVolume()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		pins = append(pins, placementservice.VolumePin{
			VolumeName: setting.Name,
			NodeID:     vol.NodeID,
			NodeLabel:  vol.NodeLabel,
		})
	}
	return pins, nil
}
