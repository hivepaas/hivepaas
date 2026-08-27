package placementserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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

	return nil
}
