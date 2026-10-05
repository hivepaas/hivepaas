package appmetricshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmetricsuc"
)

type Handler struct {
	*appbasehandler.Handler
	appMetricsUC *appmetricsuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appMetricsUC *appmetricsuc.UC,
) *Handler {
	return &Handler{
		Handler:      baseHandler,
		appMetricsUC: appMetricsUC,
	}
}
