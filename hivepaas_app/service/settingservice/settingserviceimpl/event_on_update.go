package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

func (s *service) OnUpdate(
	ctx context.Context,
	db database.IDB,
	event *settingservice.UpdateEvent,
) (err error) {
	// Remove periodic jobs cache as the update may relate
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		err = s.periodicSettingsRepo.Del(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
