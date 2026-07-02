package appcopyserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) applyAppHttpSettings(
	ctx context.Context,
	data *appCopyData,
) error {
	app := data.TargetApp
	httpSetting := app.GetSettingByType(base.SettingTypeAppHttp)
	httpSettings, err := httpSetting.AsAppHttpSettings()
	if err != nil {
		return apperrors.New(err)
	}

	mapSslSettings := map[string]*entity.Setting{}
	for _, sslID := range httpSettings.GetSSLCertIDs() {
		if s := data.RefObjects.RefSettings[sslID]; s != nil {
			mapSslSettings[s.ID] = s
		}
	}
	err = s.sslService.WriteCertFiles(false, gofn.MapValues(mapSslSettings)...)
	if err != nil {
		return apperrors.New(err)
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	service := &inspect.Service

	err = s.traefikService.ApplyAppConfig(ctx, app, service, &traefikservice.AppConfigData{
		HttpSettings: httpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.New(err)
	}

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
