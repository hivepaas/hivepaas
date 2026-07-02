package appactionhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc"
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
