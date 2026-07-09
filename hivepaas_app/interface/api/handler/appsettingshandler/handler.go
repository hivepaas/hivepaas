package appsettingshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc"
)

type Handler struct {
	*appbasehandler.Handler
	appSettingsUC *appsettingsuc.UC
	fileUC        *fileuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appSettingsUC *appsettingsuc.UC,
	fileUC *fileuc.UC,
) *Handler {
	return &Handler{
		Handler:       baseHandler,
		appSettingsUC: appSettingsUC,
		fileUC:        fileUC,
	}
}
