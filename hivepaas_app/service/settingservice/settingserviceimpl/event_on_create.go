package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

func (s *service) OnCreate(
	_ context.Context,
	_ database.IDB,
	_ *settingservice.CreateEvent,
) (err error) {
	return nil
}
