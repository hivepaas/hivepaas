package appactionhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appbasehandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appactionuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
)

type Handler struct {
	*appbasehandler.Handler
	appUC       *appuc.UC
	appActionUC *appactionuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appUC *appuc.UC,
	appActionUC *appactionuc.UC,
) *Handler {
	return &Handler{
		Handler:     baseHandler,
		appUC:       appUC,
		appActionUC: appActionUC,
	}
}
