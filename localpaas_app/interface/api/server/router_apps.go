package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerAppRoutes(projectGroup *gin.RouterGroup) *gin.RouterGroup {
	appGroup := projectGroup.Group("/:projectID/apps")
	appHandler := s.handlerRegistry.appHandler
	appSettingsHandler := s.handlerRegistry.appSettingsHandler
	appDeploymentHandler := s.handlerRegistry.appDeploymentHandler

	// Info
	appGroup.GET("/base", appHandler.ListAppBase)
	appGroup.GET("/:appID", appHandler.GetApp)
	appGroup.GET("", appHandler.ListApp)
	// Creation & Update
	appGroup.POST("", appHandler.CreateApp)
	appGroup.PUT("/:appID", appHandler.UpdateApp)
	appGroup.DELETE("/:appID", appHandler.DeleteApp)
	// Token
	appGroup.PUT("/:appID/token", appHandler.UpdateAppToken)
	// Logs
	appGroup.GET("/:appID/runtime-logs", func(ctx *gin.Context) {
		appHandler.GetAppRuntimeLogs(ctx, s.websocket)
	})

	// Tags
	appGroup.POST("/:appID/tags", appSettingsHandler.CreateAppTag)
	appGroup.POST("/:appID/tags/delete", appSettingsHandler.DeleteAppTags)
	// Settings
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
	// Env vars
	appGroup.GET("/:appID/env-vars", appSettingsHandler.GetAppEnvVars)
	appGroup.PUT("/:appID/env-vars", appSettingsHandler.UpdateAppEnvVars)
	// Secrets
	appGroup.GET("/:appID/secrets", appSettingsHandler.ListAppSecret)
	appGroup.POST("/:appID/secrets", appSettingsHandler.CreateAppSecret)
	appGroup.PUT("/:appID/secrets/:itemID", appSettingsHandler.UpdateAppSecret)
	appGroup.PUT("/:appID/secrets/:itemID/status", appSettingsHandler.UpdateAppSecretStatus)
	appGroup.DELETE("/:appID/secrets/:itemID", appSettingsHandler.DeleteAppSecret)

	cronJobGroup := appGroup.Group("/:appID/cron-jobs")
	{ // Cron job group
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

	healthcheckGroup := appGroup.Group("/:appID/healthchecks")
	{ // healthcheck group
		healthcheckGroup.GET("", appSettingsHandler.ListAppHealthcheck)
		healthcheckGroup.GET("/:itemID", appSettingsHandler.GetAppHealthcheck)
		healthcheckGroup.POST("", appSettingsHandler.CreateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID", appSettingsHandler.UpdateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID/status", appSettingsHandler.UpdateAppHealthcheckStatus)
		healthcheckGroup.DELETE("/:itemID", appSettingsHandler.DeleteAppHealthcheck)

		// Healthcheck task group
		healthcheckGroup.GET("/:itemID/tasks", appSettingsHandler.ListAppHealthcheckTask)
	}

	appContainerGroup := appGroup.Group("/:appID/container")
	{ // app container group
		// Cancel
		appContainerGroup.POST("/check-port", appSettingsHandler.CheckAppContainerPort)
	}

	appDeploymentGroup := appGroup.Group("/:appID/deployments")
	{ // app deployment group
		// Info
		appDeploymentGroup.GET("/:deploymentID", appDeploymentHandler.GetAppDeployment)
		appDeploymentGroup.GET("", appDeploymentHandler.ListAppDeployment)
		// Cancel
		appDeploymentGroup.POST("/:deploymentID/cancel", appDeploymentHandler.CancelAppDeployment)
		// Logs
		appDeploymentGroup.GET("/:deploymentID/logs", func(ctx *gin.Context) {
			appDeploymentHandler.GetAppDeploymentLogs(ctx, s.websocket)
		})
	}

	return appGroup
}
