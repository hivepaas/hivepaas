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
	// How a request reaches this install - used to work out the proxy settings
	hivepaasGroup.GET("/request-info", hivepaasHandler.GetRequestInfo)

	// Release info
	hivepaasGroup.GET("/release-info", hivepaasHandler.GetAppReleaseInfo)
	// Update app version
	hivepaasGroup.POST("/update-version", hivepaasHandler.UpdateAppVersion)

	// Service settings
	hivepaasGroup.GET("/service-settings", hivepaasHandler.GetServiceSettings)
	hivepaasGroup.PUT("/service-settings", hivepaasHandler.UpdateServiceSettings)

	// App secret
	hivepaasGroup.PUT("/app-secret", hivepaasHandler.UpdateAppSecret)

	// Routing settings
	hivepaasGroup.GET("/routing-settings", hivepaasHandler.GetRoutingSettings)
	hivepaasGroup.PUT("/routing-settings", hivepaasHandler.UpdateRoutingSettings)
	// Confirm-or-revert. A routing change applied through the PUT above is undone
	// at its deadline unless a call gets back in here to vouch for it.
	hivepaasGroup.POST("/routing-settings/confirm", hivepaasHandler.ConfirmRoutingSettings)
	hivepaasGroup.POST("/routing-settings/revert", hivepaasHandler.RevertRoutingSettings)
	hivepaasGroup.POST("/service-settings/confirm", hivepaasHandler.ConfirmServiceSettings)
	hivepaasGroup.POST("/service-settings/revert", hivepaasHandler.RevertServiceSettings)

	// Security settings
	hivepaasGroup.GET("/security-settings", hivepaasHandler.GetSecuritySettings)
	hivepaasGroup.PUT("/security-settings", hivepaasHandler.UpdateSecuritySettings)

	return hivepaasGroup
}
