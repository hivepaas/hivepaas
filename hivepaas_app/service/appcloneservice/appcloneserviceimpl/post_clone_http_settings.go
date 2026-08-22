package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

func (s *service) applyAppHttpSettings(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	app := data.DestApp
	httpSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	if httpSetting == nil {
		return nil
	}
	httpSettings, err := httpSetting.AsAppRoutingSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = s.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             app,
		RoutingSettings: httpSettings,
		RefObjects:      data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
