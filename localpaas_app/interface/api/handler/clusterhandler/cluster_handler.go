package clusterhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc"
)

type ClusterHandler struct {
	*handler.BaseHandler
	authHandler *authhandler.AuthHandler
	nodeUC      *nodeuc.NodeUC
	volumeUC    *volumeuc.VolumeUC
	imageUC     *imageuc.ImageUC
	networkUC   *networkuc.NetworkUC
}

func NewClusterHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	nodeUC *nodeuc.NodeUC,
	volumeUC *volumeuc.VolumeUC,
	imageUC *imageuc.ImageUC,
	networkUC *networkuc.NetworkUC,
) *ClusterHandler {
	return &ClusterHandler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		nodeUC:      nodeUC,
		volumeUC:    volumeUC,
		imageUC:     imageUC,
		networkUC:   networkUC,
	}
}
