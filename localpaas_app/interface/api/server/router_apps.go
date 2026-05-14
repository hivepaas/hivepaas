package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerAppRoutes(projectGroup *gin.RouterGroup) *gin.RouterGroup {
	appGroup := projectGroup.Group("/:projectID/apps")
	appHandler := s.handlerRegistry.appHandler
	appSettingsHandler := s.handlerRegistry.appSettingsHandler
	appDeploymentHandler := s.handlerRegistry.appDeploymentHandler

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
		// Token
		appGroup.PUT("/:appID/token", appHandler.UpdateAppToken)
		// Logs
		appGroup.GET("/:appID/runtime-logs", func(ctx *gin.Context) {
			appHandler.GetAppRuntimeLogs(ctx, s.websocket)
		})
	}

	{ // Tags
		tagGroup := appGroup.Group("/:appID/tags")
		tagGroup.POST("", appSettingsHandler.CreateAppTag)
		tagGroup.POST("/delete", appSettingsHandler.DeleteAppTags)
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
		appGroup.GET("/:appID/container-settings", appSettingsHandler.GetAppContainerSettings)
		appGroup.PUT("/:appID/container-settings", appSettingsHandler.UpdateAppContainerSettings)
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

	{ // Cron jobs
		cronJobGroup := appGroup.Group("/:appID/cron-jobs")
		cronJobGroup.GET("", appSettingsHandler.ListAppCronJob)
		cronJobGroup.GET("/:itemID", appSettingsHandler.GetAppCronJob)
		cronJobGroup.POST("", appSettingsHandler.CreateAppCronJob)
		cronJobGroup.PUT("/:itemID", appSettingsHandler.UpdateAppCronJob)
		cronJobGroup.PUT("/:itemID/status", appSettingsHandler.UpdateAppCronJobStatus)
		cronJobGroup.DELETE("/:itemID", appSettingsHandler.DeleteAppCronJob)
		// Execute
		cronJobGroup.POST("/:itemID/exec", appSettingsHandler.ExecuteAppCronJob)

		// Cron job task group
		cronJobGroup.GET("/:itemID/tasks", appSettingsHandler.ListAppCronJobTask)
		cronJobGroup.GET("/:itemID/tasks/:taskID", appSettingsHandler.GetAppCronJobTask)
		cronJobGroup.GET("/:itemID/tasks/:taskID/logs", func(ctx *gin.Context) {
			appSettingsHandler.GetAppCronJobTaskLogs(ctx, s.websocket)
		})
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
		// Cancel
		deploymentGroup.POST("/:deploymentID/cancel", appDeploymentHandler.CancelAppDeployment)
		// Logs
		deploymentGroup.GET("/:deploymentID/logs", func(ctx *gin.Context) {
			appDeploymentHandler.GetAppDeploymentLogs(ctx, s.websocket)
		})
	}

	return appGroup
}
