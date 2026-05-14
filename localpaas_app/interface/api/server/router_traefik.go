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

	return traefikGroup
}
