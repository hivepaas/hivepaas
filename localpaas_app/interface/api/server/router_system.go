package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSystemRoutes(apiGroup *gin.RouterGroup) {
	systemGroup := apiGroup.Group("/system")
	systemHandler := s.handlerRegistry.systemHandler

	{ // task group
		taskGroup := systemGroup.Group("/tasks")
		taskGroup.GET("", systemHandler.ListTask)
		taskGroup.GET("/:taskID", systemHandler.GetTask)
		taskGroup.GET("/:taskID/status", systemHandler.GetTaskStatus)
		taskGroup.POST("/:taskID/cancel", systemHandler.CancelTask)
	}

	{ // error group
		errorGroup := systemGroup.Group("/errors")
		errorGroup.GET("", systemHandler.ListSysError)
		errorGroup.GET("/:errorID", systemHandler.GetSysError)
		errorGroup.DELETE("/:errorID", systemHandler.DeleteSysError)
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

	{ // Backup group
		backupGroup := systemSettingGroup.Group("/backup")
		backupGroup.GET("", systemSettingsHandler.GetBackupSettings)
		backupGroup.PUT("", systemSettingsHandler.UpdateBackupSettings)
		backupGroup.POST("/exec", systemSettingsHandler.ExecuteBackup)

		// Backup files
		backupGroup.GET("/files", systemSettingsHandler.ListBackupFiles)
		backupGroup.GET("/files/:fileID", systemSettingsHandler.GetBackupFile)
		backupGroup.GET("/files/:fileID/download", systemSettingsHandler.DownloadBackupFile)
	}

	{ // SSL renewal group
		sslRenewalGroup := systemSettingGroup.Group("/ssl-renewal")
		sslRenewalGroup.GET("", systemSettingsHandler.GetSSLRenewalSettings)
		sslRenewalGroup.PUT("", systemSettingsHandler.UpdateSSLRenewalSettings)
		sslRenewalGroup.POST("/exec", systemSettingsHandler.ExecuteSSLRenewal)
	}

	_ = s.registerLocalPaaSRoutes(systemGroup)
	_ = s.registerTraefikRoutes(systemGroup)
}
