package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

func (s *service) OnUpdateStatus(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.UpdateEvent,
) (err error) {
	// Reload periodic jobs in workers as the update may relate
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	}

	return nil
}
