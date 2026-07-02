package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerHivePaaSRoutes(systemGroup *gin.RouterGroup) *gin.RouterGroup {
	hivepaasGroup := systemGroup.Group("/hivepaas")
	hivepaasHandler := s.handlerRegistry.hivepaasHandler

	// Process
	hivepaasGroup.POST("/restart", hivepaasHandler.RestartHivePaaSApp)
	// Config
	hivepaasGroup.POST("/config/reload", hivepaasHandler.ReloadHivePaaSAppConfig)

	// Release info
	hivepaasGroup.GET("/release-info", hivepaasHandler.GetAppReleaseInfo)
	// Update app version
	hivepaasGroup.POST("/update-version", hivepaasHandler.UpdateAppVersion)

	// Service settings
	hivepaasGroup.GET("/service-settings", hivepaasHandler.GetServiceSettings)
	hivepaasGroup.PUT("/service-settings", hivepaasHandler.UpdateServiceSettings)

	return hivepaasGroup
}
