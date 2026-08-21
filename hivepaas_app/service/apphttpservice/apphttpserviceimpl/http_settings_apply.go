package apphttpserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

const (
	defaultServiceUpdateRetryMax   = 2
	defaultServiceUpdateRetryDelay = time.Second * 2
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

	if !data.SkipUpdatingService {
		err = s.dockerManager.ServiceUpdateFunc(ctx, appSvc.ID,
			func(_ int, svc *swarm.Service) error {
				if !data.SkipApplyingNetworks {
					if err := s.networkService.UpdateAppGlobalRoutingNetwork(ctx, app, svc, data.HttpSettings); err != nil {
						return apperrors.Wrap(err)
					}
				}
				return nil
			}, defaultServiceUpdateRetryMax, defaultServiceUpdateRetryDelay, 0)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else if !data.SkipApplyingNetworks {
		err = s.networkService.UpdateAppGlobalRoutingNetwork(ctx, app, appSvc, data.HttpSettings)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
