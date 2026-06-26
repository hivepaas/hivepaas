package appcopyserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) applySwarmConfigFiles(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	app := data.TargetApp
	configSettings := app.GetSettingsByType(base.SettingTypeConfigFile)
	if len(configSettings) == 0 {
		return nil
	}
	configItems := make([]*entity.ConfigFile, 0, len(configSettings))
	for _, configItem := range configSettings {
		configItems = append(configItems, configItem.MustAsConfigFile())
	}
	data.TargetConfig, err = s.clusterService.CreateConfigsForApp(ctx, db, app, configItems)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
