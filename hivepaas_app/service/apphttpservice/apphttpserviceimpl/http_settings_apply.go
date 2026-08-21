package apphttpserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) applyHttpSettings(
	ctx context.Context,
	db database.IDB,
	data *applyAppHttpData,
) (err error) {
	app := data.App
	appSvc := data.Service
	appHttpSettings := data.HttpSettings

	if !data.SkipApplyingSslCerts {
		mapSslSettings := map[string]*entity.Setting{}
		for _, sslID := range appHttpSettings.GetSSLCertIDs() {
			if s := data.RefObjects.RefSettings[sslID]; s != nil {
				mapSslSettings[s.ID] = s
			}
		}
		err = s.sslService.WriteCertFiles(data.ForceRecreateSslCertFiles, gofn.MapValues(mapSslSettings)...)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	_, err = s.traefikService.ApplyAppConfig(ctx, db, &traefikservice.ApplyAppConfigReq{
		App:          app,
		Service:      appSvc,
		HttpSettings: appHttpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	if !data.SkipApplyingNetworks {
		err = s.networkService.UpdateAppGlobalRoutingNetwork(ctx, app, appSvc, data.HttpSettings)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	if !data.SkipUpdatingService {
		// NOTE: don't use ServiceUpdateFunc in this context
		_, err = s.dockerManager.ServiceUpdate(ctx, appSvc.ID, &appSvc.Version, &appSvc.Spec)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
