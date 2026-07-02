package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerAppRoutes(projectGroup *gin.RouterGroup) *gin.RouterGroup {
	appGroup := projectGroup.Group("/:projectID/apps")
	appHandler := s.handlerRegistry.appHandler
	appSettingsHandler := s.handlerRegistry.appSettingsHandler
	appDeploymentHandler := s.handlerRegistry.appDeploymentHandler
	appActionHandler := s.handlerRegistry.appActionHandler
	appPreviewHandler := s.handlerRegistry.appPreviewHandler

	{ // Base
		appGroup.GET("/base", appHandler.ListAppBase)
		appGroup.GET("/:appID", appHandler.GetApp)
		appGroup.GET("", appHandler.ListApp)
		// Creation & Update
		appGroup.POST("", appHandler.CreateApp)
		appGroup.PUT("/:appID", appHandler.UpdateApp)
		appGroup.DELETE("/:appID", appHandler.DeleteApp)
		// Status Update
		appGroup.PUT("/:appID/status", appHandler.UpdateAppStatus)
	}

	{ // Copy
		appGroup.GET("/:appID/copy/prepare", appHandler.PrepareAppCopy)
		appGroup.POST("/:appID/copy", appHandler.CopyApp)
	}

	{ // Tags
		tagGroup := appGroup.Group("/:appID/tags")
		tagGroup.POST("", appSettingsHandler.CreateAppTag)
		tagGroup.POST("/delete", appSettingsHandler.DeleteAppTags)
	}

	{ // Logs
		appGroup.GET("/:appID/logs/info", appHandler.GetAppLogsInfo)
		appGroup.GET("/:appID/logs", appHandler.GetAppLogs)
	}

	{ // Terminal
		appGroup.GET("/:appID/terminal/info", appHandler.GetAppTerminalInfo)
		appGroup.GET("/:appID/terminal", appHandler.OpenAppTerminal)
	}

	{ // Settings
		appGroup.GET("/:appID/deployment-settings", appSettingsHandler.GetAppDeploymentSettings)
		appGroup.PUT("/:appID/deployment-settings", appSettingsHandler.UpdateAppDeploymentSettings)
		appGroup.GET("/:appID/http-settings", appSettingsHandler.GetAppHttpSettings)
		appGroup.PUT("/:appID/http-settings", appSettingsHandler.UpdateAppHttpSettings)
		appGroup.GET("/:appID/service-settings", appSettingsHandler.GetAppServiceSettings)
		appGroup.PUT("/:appID/service-settings", appSettingsHandler.UpdateAppServiceSettings)
		appGroup.GET("/:appID/network-settings", appSettingsHandler.GetAppNetworkSettings)
		appGroup.PUT("/:appID/network-settings", appSettingsHandler.UpdateAppNetworkSettings)
		appGroup.GET("/:appID/resource-settings", appSettingsHandler.GetAppResourceSettings)
		appGroup.PUT("/:appID/resource-settings", appSettingsHandler.UpdateAppResourceSettings)
		appGroup.GET("/:appID/storage-settings", appSettingsHandler.GetAppStorageSettings)
		appGroup.PUT("/:appID/storage-settings", appSettingsHandler.UpdateAppStorageSettings)
		appGroup.GET("/:appID/feature-settings", appSettingsHandler.GetAppFeatureSettings)
		appGroup.PUT("/:appID/feature-settings", appSettingsHandler.UpdateAppFeatureSettings)
		appGroup.GET("/:appID/container-settings", appSettingsHandler.GetAppContainerSettings)
		appGroup.PUT("/:appID/container-settings", appSettingsHandler.UpdateAppContainerSettings)
	}

	{ // Service tasks
		appGroup.GET("/:appID/service-tasks", appSettingsHandler.GetAppServiceTasks)
	}

	{ // Env vars
		envVarGroup := appGroup.Group("/:appID/env-vars")
		envVarGroup.GET("", appSettingsHandler.GetEnvVars)
		envVarGroup.PUT("", appSettingsHandler.UpdateEnvVars)
	}

	{ // Secrets
		secretGroup := appGroup.Group("/:appID/secrets")
		secretGroup.GET("", appSettingsHandler.ListSecret)
		secretGroup.GET("/:itemID", appSettingsHandler.GetSecret)
		secretGroup.POST("", appSettingsHandler.CreateSecret)
		secretGroup.PUT("/:itemID", appSettingsHandler.UpdateSecret)
		secretGroup.PUT("/:itemID/status", appSettingsHandler.UpdateSecretStatus)
		secretGroup.DELETE("/:itemID", appSettingsHandler.DeleteSecret)
		// Download as file
		secretGroup.GET("/:itemID/download-token", appSettingsHandler.GetSecretDownloadToken)
		secretGroup.GET("/:itemID/download", appSettingsHandler.DownloadSecret)
	}

	{ // Config files
		configFileGroup := appGroup.Group("/:appID/config-files")
		configFileGroup.GET("", appSettingsHandler.ListConfigFile)
		configFileGroup.GET("/:itemID", appSettingsHandler.GetConfigFile)
		configFileGroup.POST("", appSettingsHandler.CreateConfigFile)
		configFileGroup.PUT("/:itemID", appSettingsHandler.UpdateConfigFile)
		configFileGroup.PUT("/:itemID/status", appSettingsHandler.UpdateConfigFileStatus)
		configFileGroup.DELETE("/:itemID", appSettingsHandler.DeleteConfigFile)
		// Download as file
		configFileGroup.GET("/:itemID/download-token", appSettingsHandler.GetConfigFileDownloadToken)
		configFileGroup.GET("/:itemID/download", appSettingsHandler.DownloadConfigFile)
	}

	{ // Scheduled jobs
		schedJobGroup := appGroup.Group("/:appID/sched-jobs")
		schedJobGroup.GET("", appSettingsHandler.ListAppSchedJob)
		schedJobGroup.GET("/:itemID", appSettingsHandler.GetAppSchedJob)
		schedJobGroup.POST("", appSettingsHandler.CreateAppSchedJob)
		schedJobGroup.PUT("/:itemID", appSettingsHandler.UpdateAppSchedJob)
		schedJobGroup.PUT("/:itemID/status", appSettingsHandler.UpdateAppSchedJobStatus)
		schedJobGroup.DELETE("/:itemID", appSettingsHandler.DeleteAppSchedJob)
		// Execute
		schedJobGroup.POST("/:itemID/exec", appSettingsHandler.ExecuteAppSchedJob)

		// Sched job task group
		schedJobGroup.GET("/:itemID/tasks", appSettingsHandler.ListAppSchedJobTask)
		schedJobGroup.GET("/:itemID/tasks/:taskID", appSettingsHandler.GetAppSchedJobTask)
		schedJobGroup.POST("/:itemID/tasks/:taskID/cancel", appSettingsHandler.CancelAppSchedJobTask)
		schedJobGroup.GET("/:itemID/tasks/:taskID/logs", appSettingsHandler.GetAppSchedJobTaskLogs)
	}

	{ // Health checks
		healthcheckGroup := appGroup.Group("/:appID/healthchecks")
		healthcheckGroup.GET("", appSettingsHandler.ListAppHealthcheck)
		healthcheckGroup.GET("/:itemID", appSettingsHandler.GetAppHealthcheck)
		healthcheckGroup.POST("", appSettingsHandler.CreateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID", appSettingsHandler.UpdateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID/status", appSettingsHandler.UpdateAppHealthcheckStatus)
		healthcheckGroup.DELETE("/:itemID", appSettingsHandler.DeleteAppHealthcheck)

		// Healthcheck task group
		healthcheckGroup.GET("/:itemID/tasks", appSettingsHandler.ListAppHealthcheckTask)
	}

	{ // App containers
		containerGroup := appGroup.Group("/:appID/container")
		// Check port
		containerGroup.POST("/check-port", appSettingsHandler.CheckAppContainerPort)
	}

	{ // Deployments
		deploymentGroup := appGroup.Group("/:appID/deployments")
		// Info
		deploymentGroup.GET("/:deploymentID", appDeploymentHandler.GetAppDeployment)
		deploymentGroup.GET("", appDeploymentHandler.ListAppDeployment)
		deploymentGroup.GET("/:deploymentID/status", appDeploymentHandler.GetAppDeploymentStatus)
		// Cancel
		deploymentGroup.POST("/:deploymentID/cancel", appDeploymentHandler.CancelAppDeployment)
		// Logs
		deploymentGroup.GET("/:deploymentID/logs", appDeploymentHandler.GetAppDeploymentLogs)
	}

	{ // Actions
		appActionGroup := appGroup.Group("/:appID")
		// Re-deploy app
		appActionGroup.POST("/deploy", appActionHandler.DeployApp)
		// Restart app
		appActionGroup.POST("/restart", appActionHandler.RestartApp)
	}

	{ // App previews
		previewGroup := appGroup.Group("/:appID/previews")
		previewGroup.GET("", appPreviewHandler.ListAppPreview)
		previewGroup.POST("", appPreviewHandler.CreateAppPreview)
		previewGroup.POST("/prepare", appPreviewHandler.PrepareCreateAppPreview)
	}

	return appGroup
}
