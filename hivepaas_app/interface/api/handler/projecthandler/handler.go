package projecthandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc"
)

type Handler struct {
	*projectbasehandler.Handler
	projectUC        *projectuc.UC
	clusterNetworkUC *networkuc.UC
	clusterVolumeUC  *volumeuc.UC
}

func New(
	baseHandler *projectbasehandler.Handler,
	projectUC *projectuc.UC,
	clusterNetworkUC *networkuc.UC,
	clusterVolumeUC *volumeuc.UC,
) *Handler {
	return &Handler{
		Handler:          baseHandler,
		projectUC:        projectUC,
		clusterNetworkUC: clusterNetworkUC,
		clusterVolumeUC:  clusterVolumeUC,
	}
}
