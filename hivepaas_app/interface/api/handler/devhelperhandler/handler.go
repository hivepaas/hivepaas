package devhelperhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	devHelperUC *devhelperuc.UC
	sessionUC   *sessionuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	devHelperUC *devhelperuc.UC,
	sessionUC *sessionuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		devHelperUC: devHelperUC,
		sessionUC:   sessionUC,
	}
}
