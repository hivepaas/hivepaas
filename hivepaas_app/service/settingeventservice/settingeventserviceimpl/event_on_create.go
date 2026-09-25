package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

func (s *service) OnCreate(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.CreateEvent,
) (err error) {
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	}
	return s.recordMountRefresh(ctx, db, &event.Tasks, event.Setting)
}
