package apphandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc"
)

type AppHandler struct {
	*basesettinghandler.BaseSettingHandler
	appUC           *appuc.UC
	appDeploymentUC *appdeploymentuc.UC
	secretUC        *secretuc.UC
}

func NewAppHandler(
	baseSettingHandler *basesettinghandler.BaseSettingHandler,
	appUC *appuc.UC,
	appDeploymentUC *appdeploymentuc.UC,
	secretUC *secretuc.UC,
) *AppHandler {
	return &AppHandler{
		BaseSettingHandler: baseSettingHandler,
		appUC:              appUC,
		appDeploymentUC:    appDeploymentUC,
		secretUC:           secretUC,
	}
}
