package apptemplatehandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc"
)

type Handler struct {
	*appbasehandler.Handler
	appTemplateUC *apptemplateuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appTemplateUC *apptemplateuc.UC,
) *Handler {
	return &Handler{
		Handler:       baseHandler,
		appTemplateUC: appTemplateUC,
	}
}
