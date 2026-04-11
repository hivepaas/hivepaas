package projectsettingshandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projectbasehandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectsettingsuc"
)

type Handler struct {
	*projectbasehandler.Handler
	projectSettingsUC *projectsettingsuc.UC
	dockerNetworkUC   *networkuc.UC
	dockerVolumeUC    *volumeuc.UC
}

func New(
	baseHandler *projectbasehandler.Handler,
	projectSettingsUC *projectsettingsuc.UC,
	dockerNetworkUC *networkuc.UC,
	dockerVolumeUC *volumeuc.UC,
) *Handler {
	return &Handler{
		Handler:           baseHandler,
		projectSettingsUC: projectSettingsUC,
		dockerNetworkUC:   dockerNetworkUC,
		dockerVolumeUC:    dockerVolumeUC,
	}
}
