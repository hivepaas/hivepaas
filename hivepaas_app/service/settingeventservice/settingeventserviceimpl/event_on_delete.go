package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

func (s *service) OnDelete(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.DeleteEvent,
) (err error) {
	// Reload periodic jobs in workers as the update may relate
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		if event.Setting.Type == base.SettingTypePeriodicJob {
			_ = s.periodicSettingsRepo.RemoveJob(ctx, event.Setting.ID)
		}
		_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	}

	return nil
}
