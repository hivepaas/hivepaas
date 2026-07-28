package projectenvsettingshandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc"
)

type Handler struct {
	*basesettinghandler.Handler
	projectSettingsUC *projectsettingsuc.UC
	clusterNetworkUC  *networkuc.UC
	clusterVolumeUC   *volumeuc.UC
}

func New(
	handler *basesettinghandler.Handler,
	projectSettingsUC *projectsettingsuc.UC,
	clusterNetworkUC *networkuc.UC,
	clusterVolumeUC *volumeuc.UC,
) *Handler {
	return &Handler{
		Handler:           handler,
		projectSettingsUC: projectSettingsUC,
		clusterNetworkUC:  clusterNetworkUC,
		clusterVolumeUC:   clusterVolumeUC,
	}
}
