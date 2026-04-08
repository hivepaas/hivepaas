package settingserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
)

func (s *service) OnDelete(
	ctx context.Context,
	db database.IDB,
	event *settingservice.DeleteEvent,
) (err error) {
	// Remove healthcheck cache if the update may relate
	if event.Setting.IsTypeIn(base.SettingTypeHealthcheck, base.SettingTypeIMService, base.SettingTypeEmail) {
		err = s.healthcheckSettingsRepo.Del(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
