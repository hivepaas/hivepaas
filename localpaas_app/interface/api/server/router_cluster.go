package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerClusterRoutes(apiGroup *gin.RouterGroup) {
	clusterGroup := apiGroup.Group("/cluster")
	clusterHandler := s.handlerRegistry.clusterHandler

	{ // node group
		nodeGroup := clusterGroup.Group("/nodes")
		// Nodes
		nodeGroup.GET("", clusterHandler.ListNode)
		nodeGroup.GET("/:nodeID", clusterHandler.GetNode)
		nodeGroup.GET("/:nodeID/inspect", clusterHandler.GetNodeInspection)
		nodeGroup.PUT("/:nodeID", clusterHandler.UpdateNode)
		nodeGroup.DELETE("/:nodeID", clusterHandler.DeleteNode)
		// Node join
		nodeGroup.POST("/join", clusterHandler.JoinNode)
		nodeGroup.GET("/join-command", clusterHandler.GetNodeJoinCommand)
	}

	{ // volume group
		volumeGroup := clusterGroup.Group("/volumes")
		// Volumes
		volumeGroup.GET("", clusterHandler.ListVolume)
		volumeGroup.GET("/:volumeID", clusterHandler.GetVolume)
		volumeGroup.GET("/:volumeID/inspect", clusterHandler.GetVolumeInspection)
		volumeGroup.POST("", clusterHandler.CreateVolume)
		volumeGroup.DELETE("/:volumeID", clusterHandler.DeleteVolume)
	}

	{ // image group
		imageGroup := clusterGroup.Group("/images")
		// Volumes
		imageGroup.GET("", clusterHandler.ListImage)
		imageGroup.GET("/:imageID", clusterHandler.GetImage)
		imageGroup.GET("/:imageID/inspect", clusterHandler.GetImageInspection)
		imageGroup.POST("", clusterHandler.CreateImage)
		imageGroup.DELETE("/:imageID", clusterHandler.DeleteImage)
	}

	{ // network group
		networkGroup := clusterGroup.Group("/networks")
		// Networks
		networkGroup.GET("", clusterHandler.ListNetwork)
		networkGroup.GET("/:networkID", clusterHandler.GetNetwork)
		networkGroup.GET("/:networkID/inspect", clusterHandler.GetNetworkInspection)
		networkGroup.POST("", clusterHandler.CreateNetwork)
		networkGroup.DELETE("/:networkID", clusterHandler.DeleteNetwork)
	}

	{ // build group
		buildGroup := clusterGroup.Group("/build")
		buildGroup.POST("/cache-clear", clusterHandler.ClearBuildCache)
	}
}
