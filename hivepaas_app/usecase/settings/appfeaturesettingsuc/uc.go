package appfeaturesettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType = base.SettingTypeAppFeatures
)

type UC struct {
	*settings.BaseUC

	obiSettings cacherepository.OBISettingsRepo
}

func New(
	baseUC *settings.BaseUC,
	obiSettings cacherepository.OBISettingsRepo,
) *UC {
	return &UC{
		BaseUC: baseUC,

		obiSettings: obiSettings,
	}
}

// forgetOBISettings drops the settings the agents cache to run OBI, once a
// change to an app's feature settings is committed: whether an app asks for
// its routes and calls is among them. A failure costs at most the agents'
// next read of the database, within 10 minutes.
func (uc *UC) forgetOBISettings(ctx context.Context) {
	_ = uc.obiSettings.Invalidate(context.WithoutCancel(ctx))
}
