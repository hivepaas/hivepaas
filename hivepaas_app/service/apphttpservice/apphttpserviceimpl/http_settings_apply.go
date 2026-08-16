package apphttpserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

const (
	defaultServiceUpdateRetryMax   = 2
	defaultServiceUpdateRetryDelay = time.Second * 2
)

func (s *service) applyHttpSettings(
	ctx context.Context,
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

	err = s.traefikService.ApplyAppConfig(ctx, app, appSvc, &traefikservice.AppConfigData{
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
		err = gofn.ExecRetry(func() error {
			_, err = s.dockerManager.ServiceUpdate(ctx, appSvc.ID, &appSvc.Version, &appSvc.Spec)
			return apperrors.Wrap(err)
		}, defaultServiceUpdateRetryMax, defaultServiceUpdateRetryDelay)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
