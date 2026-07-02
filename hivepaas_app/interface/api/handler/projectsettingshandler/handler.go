package projectsettingshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc"
)

type Handler struct {
	*projectbasehandler.Handler
	projectSettingsUC *projectsettingsuc.UC
	clusterNetworkUC  *networkuc.UC
	clusterVolumeUC   *volumeuc.UC
}

func New(
	baseHandler *projectbasehandler.Handler,
	projectSettingsUC *projectsettingsuc.UC,
	clusterNetworkUC *networkuc.UC,
	clusterVolumeUC *volumeuc.UC,
) *Handler {
	return &Handler{
		Handler:           baseHandler,
		projectSettingsUC: projectSettingsUC,
		clusterNetworkUC:  clusterNetworkUC,
		clusterVolumeUC:   clusterVolumeUC,
	}
}
