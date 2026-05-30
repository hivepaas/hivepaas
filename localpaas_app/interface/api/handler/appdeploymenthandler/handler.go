package appdeploymenthandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appbasehandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
)

type Handler struct {
	*appbasehandler.Handler
	appUC           *appuc.UC
	appDeploymentUC *appdeploymentuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appUC *appuc.UC,
	appDeploymentUC *appdeploymentuc.UC,
) *Handler {
	return &Handler{
		Handler:         baseHandler,
		appUC:           appUC,
		appDeploymentUC: appDeploymentUC,
	}
}
