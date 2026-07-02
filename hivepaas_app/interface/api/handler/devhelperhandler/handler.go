package devhelperhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	devHelperUC *devhelperuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	devHelperUC *devhelperuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		devHelperUC: devHelperUC,
	}
}
