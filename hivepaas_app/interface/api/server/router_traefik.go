package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerTraefikRoutes(systemGroup *gin.RouterGroup) *gin.RouterGroup {
	traefikGroup := systemGroup.Group("/traefik")
	traefikHandler := s.handlerRegistry.traefikHandler

	// Process
	traefikGroup.POST("/restart", traefikHandler.RestartTraefik)
	// Config
	traefikGroup.POST("/config/reload", traefikHandler.ReloadTraefikConfig)
	traefikGroup.POST("/config/reset", traefikHandler.ResetTraefikConfig)

	// Service settings
	traefikGroup.GET("/service-settings", traefikHandler.GetServiceSettings)
	traefikGroup.PUT("/service-settings", traefikHandler.UpdateServiceSettings)

	// Config options
	traefikGroup.GET("/config-options", traefikHandler.GetConfigOptions)
	traefikGroup.PUT("/config-options", traefikHandler.UpdateConfigOptions)
	// Confirm-or-revert. A command change applied through the PUT above is undone
	// unless somebody comes back through the new traefik and confirms it: an
	// argument that parses but discovers no routers leaves /ping answering 200
	// while nothing else is served, so swarm's own rollback never fires.
	traefikGroup.POST("/config-options/confirm", traefikHandler.ConfirmConfigOptions)
	traefikGroup.POST("/config-options/revert", traefikHandler.RevertConfigOptions)

	return traefikGroup
}
