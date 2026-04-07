package settingservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type UpdateEvent struct {
	Setting    *entity.Setting
	OldSetting *entity.Setting
}

func (s *settingService) OnUpdate(
	ctx context.Context,
	db database.IDB,
	event *UpdateEvent,
) (err error) {
	// Remove healthcheck cache if the update may relate
	if event.Setting.IsTypeIn(base.SettingTypeHealthcheck, base.SettingTypeIMService, base.SettingTypeEmail) {
		err = s.healthcheckSettingsRepo.Del(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Save SSL cert/key files in a directory for using later
	if event.Setting.Type == base.SettingTypeSSLCert {
		err = s.PersistSSLCertFiles(true, event.Setting)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
