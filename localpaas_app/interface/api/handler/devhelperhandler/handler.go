package devhelperhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/devhelperuc"
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
