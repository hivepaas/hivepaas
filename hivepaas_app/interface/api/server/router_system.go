package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSystemRoutes(apiGroup *gin.RouterGroup) {
	systemGroup := apiGroup.Group("/system")
	systemHandler := s.handlerRegistry.systemHandler

	{ // Task group
		taskGroup := systemGroup.Group("/tasks")
		taskGroup.GET("", systemHandler.ListTask)
		taskGroup.GET("/:itemID", systemHandler.GetTask)
		taskGroup.GET("/:itemID/status", systemHandler.GetTaskStatus)
		taskGroup.POST("/:itemID/cancel", systemHandler.CancelTask)
		taskGroup.GET("/:itemID/logs", systemHandler.GetTaskLogs)
	}

	{ // Status group
		statusGroup := systemGroup.Group("/status")
		statusGroup.GET("/db", systemHandler.GetDBStats)
	}

	{ // Error group
		errorGroup := systemGroup.Group("/errors")
		errorGroup.GET("", systemHandler.ListSysError)
		errorGroup.GET("/:errorID", systemHandler.GetSysError)
		errorGroup.DELETE("/:errorID", systemHandler.DeleteSysError)
	}

	{ // Audit log group
		auditLogGroup := systemGroup.Group("/audit-logs")
		auditLogGroup.GET("", systemHandler.ListAuditLog)
		auditLogGroup.GET("/types", systemHandler.ListAuditLogTypes)
		auditLogGroup.GET("/:itemID", systemHandler.GetAuditLog)
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
		backupGroup.DELETE("/files/:fileID", systemSettingsHandler.DeleteBackupFile)
	}

	{ // SSL renewal group
		sslRenewalGroup := systemSettingGroup.Group("/ssl-renewal")
		sslRenewalGroup.GET("", systemSettingsHandler.GetSSLRenewalSettings)
		sslRenewalGroup.PUT("", systemSettingsHandler.UpdateSSLRenewalSettings)
		sslRenewalGroup.POST("/exec", systemSettingsHandler.ExecuteSSLRenewal)
	}

	{ // Backup repo cleanup group
		backupRepoCleanupGroup := systemSettingGroup.Group("/backup-repo-cleanup")
		backupRepoCleanupGroup.GET("", systemSettingsHandler.GetBackupRepoCleanupSettings)
		backupRepoCleanupGroup.PUT("", systemSettingsHandler.UpdateBackupRepoCleanupSettings)
		backupRepoCleanupGroup.POST("/exec", systemSettingsHandler.ExecuteBackupRepoCleanup)
	}

	_ = s.registerHivePaaSRoutes(systemGroup)
	_ = s.registerTraefikRoutes(systemGroup)
}
