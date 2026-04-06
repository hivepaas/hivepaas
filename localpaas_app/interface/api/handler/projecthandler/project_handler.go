package projecthandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc"
)

type ProjectHandler struct {
	*basesettinghandler.BaseSettingHandler
	projectUC *projectuc.ProjectUC
	networkUC *networkuc.NetworkUC
	volumeUC  *volumeuc.VolumeUC
}

func NewProjectHandler(
	baseSettingHandler *basesettinghandler.BaseSettingHandler,
	projectUC *projectuc.ProjectUC,
	networkUC *networkuc.NetworkUC,
	volumeUC *volumeuc.VolumeUC,
) *ProjectHandler {
	return &ProjectHandler{
		BaseSettingHandler: baseSettingHandler,
		projectUC:          projectUC,
		networkUC:          networkUC,
		volumeUC:           volumeUC,
	}
}
