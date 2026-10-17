package approutingserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

// updateMaxFailureRatio is the share of tasks that may fail an update before it
// is rolled back. It matches what every other swarm update in the codebase sets,
// so a revert converges on the same terms as the change it undoes.
const updateMaxFailureRatio = 0.5

func (s *service) applyRoutingSettings(
	ctx context.Context,
	db database.IDB,
	data *applyAppRoutingData,
) (err error) {
	app := data.App
	appSvc := data.Service
	routingSettings := data.RoutingSettings

	if !data.SkipApplyingSslCerts {
		mapSslSettings := map[string]*entity.Setting{}
		for _, sslID := range routingSettings.GetSSLCertIDs() {
			if s := data.RefObjects.RefSettings[sslID]; s != nil {
				mapSslSettings[s.ID] = s
			}
		}
		err = s.sslService.WriteCertFiles(data.ForceRecreateSslCertFiles, gofn.MapValues(mapSslSettings)...)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	_, err = s.traefikService.ApplyAppConfig(ctx, db, &traefikservice.ApplyAppConfigReq{
		App:             app,
		Service:         appSvc,
		RoutingSettings: routingSettings,
		RefObjects:      data.RefObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	if !data.SkipApplyingNetworks {
		err = s.networkService.UpdateAppGlobalRoutingNetwork(ctx, app, appSvc, data.RoutingSettings)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	if data.RollbackOnFailure {
		setRollbackOnFailure(&appSvc.Spec)
	}

	// NOTE: don't use ServiceUpdateFunc in this context. On a retry it re-inspects
	// the service and pushes what it finds, which has none of the changes made to
	// the spec above. ApplyRoutingSettings retries the whole apply instead.
	_, err = s.dockerManager.ServiceUpdate(ctx, appSvc.ID, &appSvc.Version, &appSvc.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

// setRollbackOnFailure has swarm roll the service back when its update fails,
// rather than leave it on a configuration that does not come up.
func setRollbackOnFailure(spec *swarm.ServiceSpec) {
	if spec.UpdateConfig == nil {
		spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	spec.UpdateConfig.MaxFailureRatio = updateMaxFailureRatio
}
