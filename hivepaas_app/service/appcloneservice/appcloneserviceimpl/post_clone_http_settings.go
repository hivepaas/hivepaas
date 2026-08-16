package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apphttpservice"
)

func (s *service) applyAppHttpSettings(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	app := data.DestApp
	httpSetting := app.GetSettingByType(base.SettingTypeAppHttp)
	if httpSetting == nil {
		return nil
	}
	httpSettings, err := httpSetting.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = s.appHttpService.ApplyHttpSettings(ctx, db, &apphttpservice.ApplyAppHttpReq{
		App:          app,
		HttpSettings: httpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
