package clusterhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/builduc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc"
)

type Handler struct {
	*basesettinghandler.Handler
	nodeUC    *nodeuc.UC
	volumeUC  *volumeuc.UC
	imageUC   *imageuc.UC
	networkUC *networkuc.UC
	buildUC   *builduc.UC
}

func New(
	baseHandler *basesettinghandler.Handler,
	nodeUC *nodeuc.UC,
	volumeUC *volumeuc.UC,
	imageUC *imageuc.UC,
	networkUC *networkuc.UC,
	buildUC *builduc.UC,
) *Handler {
	return &Handler{
		Handler:   baseHandler,
		nodeUC:    nodeUC,
		volumeUC:  volumeUC,
		imageUC:   imageUC,
		networkUC: networkUC,
		buildUC:   buildUC,
	}
}
