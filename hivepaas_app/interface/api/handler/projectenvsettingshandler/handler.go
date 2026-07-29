package projectenvsettingshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc"
)

type Handler struct {
	*basesettinghandler.Handler
	projectEnvSettingsUC *projectenvsettingsuc.UC
}

func New(
	handler *basesettinghandler.Handler,
	projectEnvSettingsUC *projectenvsettingsuc.UC,
) *Handler {
	return &Handler{
		Handler:              handler,
		projectEnvSettingsUC: projectEnvSettingsUC,
	}
}
