package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

func (s *service) applyAppRoutingSettings(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	app := data.DestApp
	routingSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	if routingSetting == nil {
		return nil
	}
	routingSettings, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = s.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             app,
		RoutingSettings: routingSettings,
		RefObjects:      data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
