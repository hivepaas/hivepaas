package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerLocalPaaSRoutes(systemGroup *gin.RouterGroup) *gin.RouterGroup {
	localpaasGroup := systemGroup.Group("/localpaas")
	localpaasHandler := s.handlerRegistry.localpaasHandler

	// Process
	localpaasGroup.POST("/restart", localpaasHandler.RestartLocalPaaSApp)
	// Config
	localpaasGroup.POST("/config/reload", localpaasHandler.ReloadLocalPaaSAppConfig)

	// Release info
	localpaasGroup.GET("/release-info", localpaasHandler.GetAppReleaseInfo)
	// Update app version
	localpaasGroup.POST("/update-version", localpaasHandler.UpdateAppVersion)

	// Service settings
	localpaasGroup.GET("/service-settings", localpaasHandler.GetServiceSettings)
	localpaasGroup.PUT("/service-settings", localpaasHandler.UpdateServiceSettings)

	return localpaasGroup
}
