package systemsettingshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
)

type Handler struct {
	*basesettinghandler.Handler
}

func New(
	baseSettingHandler *basesettinghandler.Handler,
) *Handler {
	return &Handler{
		Handler: baseSettingHandler,
	}
}
