package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) applyEnvVars(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	appEnvData, err := s.buildAppEnvVars(ctx, db, true, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Apply the changes to the app and child apps
	errMap := s.envVarService.ApplyEnvVarsForApps(ctx, db, appEnvData, false, false)
	for _, e := range errMap {
		return hperrors.Wrap(e)
	}
	return nil
}

func (s *service) buildAppEnvVars(
	ctx context.Context,
	db database.IDB,
	inTx bool,
	data *appCloneData,
) (appEnvVarData []*envvarservice.AppEnvVarData, err error) {
	scope := data.DestApp.GetObjectScope()
	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	appEnvVarData, err = s.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope,
		false, nil, transaction, concurrency)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// TODO (high): check this
	// if data.BuildVarsChange {
	//	// For build phase env vars, just validate them
	//	_, err = s.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope,
	//		true, transaction, concurrency)
	//	if err != nil {
	//		return nil, hperrors.Wrap(err)
	//	}
	// }

	return appEnvVarData, nil
}
