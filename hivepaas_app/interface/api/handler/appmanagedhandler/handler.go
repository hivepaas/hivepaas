package appmanagedhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc"
)

type Handler struct {
	*appbasehandler.Handler
	appManagedUC *appmanageduc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appManagedUC *appmanageduc.UC,
) *Handler {
	return &Handler{
		Handler:      baseHandler,
		appManagedUC: appManagedUC,
	}
}
