package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

func (s *service) OnCreate(
	ctx context.Context,
	_ database.IDB,
	event *settingservice.CreateEvent,
) (err error) {
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		err = s.periodicSettingsRepo.PublishReload(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
