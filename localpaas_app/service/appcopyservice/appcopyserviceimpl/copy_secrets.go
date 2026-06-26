package appcopyserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) applySwarmSecrets(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	app := data.TargetApp
	secretSettings := app.GetSettingsByType(base.SettingTypeSecret)
	if len(secretSettings) == 0 {
		return nil
	}
	secretItems := make([]*entity.Secret, 0, len(secretSettings))
	for _, secretItem := range secretSettings {
		secretItems = append(secretItems, secretItem.MustAsSecret())
	}
	data.TargetSecrets, err = s.clusterService.CreateSecretsForApp(ctx, db, app, secretItems)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
