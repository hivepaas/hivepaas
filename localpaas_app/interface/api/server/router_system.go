package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSystemRoutes(apiGroup *gin.RouterGroup) (*gin.RouterGroup, *gin.RouterGroup) {
	systemGroup := apiGroup.Group("/system")
	systemHandler := s.handlerRegistry.systemHandler

	{ // task group
		taskGroup := systemGroup.Group("/tasks")
		taskGroup.GET("", systemHandler.ListTask)
		taskGroup.GET("/:taskID", systemHandler.GetTask)
		taskGroup.PUT("/:taskID/meta", systemHandler.UpdateTaskMeta)
		taskGroup.POST("/:taskID/cancel", systemHandler.CancelTask)
	}

	{ // error group
		errorGroup := systemGroup.Group("/errors")
		errorGroup.GET("", systemHandler.ListSysError)
		errorGroup.GET("/:errorID", systemHandler.GetSysError)
		errorGroup.DELETE("/:errorID", systemHandler.DeleteSysError)
	}

	{ // nginx group
		nginxGroup := systemGroup.Group("/nginx")
		// Process
		nginxGroup.POST("/restart", systemHandler.RestartNginx)
		// Config
		nginxGroup.POST("/config/reload", systemHandler.ReloadNginxConfig)
		nginxGroup.POST("/config/reset", systemHandler.ResetNginxConfig)
	}

	{ // localpaas app group
		lpAppGroup := systemGroup.Group("/localpaas")
		// Process
		lpAppGroup.POST("/restart", systemHandler.RestartLocalPaasApp)
		// Config
		lpAppGroup.POST("/config/reload", systemHandler.ReloadLocalPaasAppConfig)
	}

	// System settings group
	systemSettingGroup := systemGroup.Group("/settings")
	systemSettingsHandler := s.handlerRegistry.systemSettingsHandler

	{ // Cleanup group
		cleanupGroup := systemSettingGroup.Group("/cleanup")
		cleanupGroup.GET("", systemSettingsHandler.GetCleanupSettings)
		cleanupGroup.PUT("", systemSettingsHandler.UpdateCleanupSettings)
		cleanupGroup.POST("/exec", systemSettingsHandler.ExecuteCleanup)
	}

	return systemGroup, systemSettingGroup
}
