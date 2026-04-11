package projectbasehandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
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
