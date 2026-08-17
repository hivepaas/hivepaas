package appcontainerhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc"
)

type Handler struct {
	*appbasehandler.Handler
	appContainerUC *appcontaineruc.UC
	fileUC         *fileuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appContainerUC *appcontaineruc.UC,
	fileUC *fileuc.UC,
) *Handler {
	return &Handler{
		Handler:        baseHandler,
		appContainerUC: appContainerUC,
		fileUC:         fileUC,
	}
}
