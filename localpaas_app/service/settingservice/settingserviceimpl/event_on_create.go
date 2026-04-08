package settingserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
)

func (s *service) OnCreate(
	_ context.Context,
	_ database.IDB,
	_ *settingservice.CreateEvent,
) (err error) {
	return nil
}
