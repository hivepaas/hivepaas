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
		taskGroup.POST("/:taskID/cancel", systemHandler.CancelTask)
	}

	{ // error group
		errorGroup := systemGroup.Group("/errors")
		errorGroup.GET("", systemHandler.ListSysError)
		errorGroup.GET("/:errorID", systemHandler.GetSysError)
		errorGroup.DELETE("/:errorID", systemHandler.DeleteSysError)
	}

	{ // traefik group
		traefikGroup := systemGroup.Group("/traefik")
		// Process
		traefikGroup.POST("/restart", systemHandler.RestartTraefik)
		// Config
		traefikGroup.POST("/config/reload", systemHandler.ReloadTraefikConfig)
		traefikGroup.POST("/config/reset", systemHandler.ResetTraefikConfig)
	}

	{ // localpaas app group
		lpAppGroup := systemGroup.Group("/localpaas")
		// Process
		lpAppGroup.POST("/restart", systemHandler.RestartLocalPaaSApp)
		// Config
		lpAppGroup.POST("/config/reload", systemHandler.ReloadLocalPaaSAppConfig)
		// Release info
		lpAppGroup.GET("/release-info", systemHandler.GetLocalPaaSReleaseInfo)
		// Update app version
		lpAppGroup.POST("/update-version", systemHandler.UpdateLocalPaaSAppVersion)
	}

	// System settings group
	systemSettingGroup := systemGroup.Group("/settings")
	systemSettingsHandler := s.handlerRegistry.systemSettingsHandler

	{ // LocalPaaS group
		localpaasGroup := systemSettingGroup.Group("/localpaas")
		localpaasGroup.GET("", systemSettingsHandler.GetLocalPaaSSettings)
		localpaasGroup.PUT("", systemSettingsHandler.UpdateLocalPaaSSettings)
	}

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

	return systemGroup, systemSettingGroup
}
