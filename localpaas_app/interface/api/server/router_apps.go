package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerAppRoutes(projectGroup *gin.RouterGroup) *gin.RouterGroup {
	appGroup := projectGroup.Group("/:projectID/apps")
	appHandler := s.handlerRegistry.appHandler

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
	// Tags
	appGroup.POST("/:appID/tags", appHandler.CreateAppTag)
	appGroup.POST("/:appID/tags/delete", appHandler.DeleteAppTags)
	// Settings
	appGroup.GET("/:appID/deployment-settings", appHandler.GetAppDeploymentSettings)
	appGroup.PUT("/:appID/deployment-settings", appHandler.UpdateAppDeploymentSettings)
	appGroup.GET("/:appID/http-settings", appHandler.GetAppHttpSettings)
	appGroup.PUT("/:appID/http-settings", appHandler.UpdateAppHttpSettings)
	appGroup.GET("/:appID/service-settings", appHandler.GetAppServiceSettings)
	appGroup.PUT("/:appID/service-settings", appHandler.UpdateAppServiceSettings)
	appGroup.GET("/:appID/network-settings", appHandler.GetAppNetworkSettings)
	appGroup.PUT("/:appID/network-settings", appHandler.UpdateAppNetworkSettings)
	appGroup.GET("/:appID/resource-settings", appHandler.GetAppResourceSettings)
	appGroup.PUT("/:appID/resource-settings", appHandler.UpdateAppResourceSettings)
	appGroup.GET("/:appID/storage-settings", appHandler.GetAppStorageSettings)
	appGroup.PUT("/:appID/storage-settings", appHandler.UpdateAppStorageSettings)
	appGroup.GET("/:appID/container-settings", appHandler.GetAppContainerSettings)
	appGroup.PUT("/:appID/container-settings", appHandler.UpdateAppContainerSettings)
	// Env vars
	appGroup.GET("/:appID/env-vars", appHandler.GetAppEnvVars)
	appGroup.PUT("/:appID/env-vars", appHandler.UpdateAppEnvVars)
	// Secrets
	appGroup.GET("/:appID/secrets", appHandler.ListAppSecret)
	appGroup.POST("/:appID/secrets", appHandler.CreateAppSecret)
	appGroup.PUT("/:appID/secrets/:itemID", appHandler.UpdateAppSecret)
	appGroup.DELETE("/:appID/secrets/:itemID", appHandler.DeleteAppSecret)
	// Domain SSL
	appGroup.POST("/:appID/ssl/obtain", appHandler.ObtainDomainSSL)
	// Logs
	appGroup.GET("/:appID/runtime-logs", func(ctx *gin.Context) {
		appHandler.GetAppRuntimeLogs(ctx, s.websocket)
	})

	cronJobGroup := appGroup.Group("/:appID/cron-jobs")
	{ // Cron job group
		cronJobGroup.GET("", appHandler.ListAppCronJob)
		cronJobGroup.GET("/:itemID", appHandler.GetAppCronJob)
		cronJobGroup.POST("", appHandler.CreateAppCronJob)
		cronJobGroup.PUT("/:itemID", appHandler.UpdateAppCronJob)
		cronJobGroup.PUT("/:itemID/meta", appHandler.UpdateAppCronJobMeta)
		cronJobGroup.DELETE("/:itemID", appHandler.DeleteAppCronJob)
		// Execute
		cronJobGroup.POST("/:itemID/exec", appHandler.ExecuteAppCronJob)

		// Cron job task group
		cronJobGroup.GET("/:itemID/tasks", appHandler.ListAppCronJobTask)
		cronJobGroup.GET("/:itemID/tasks/:taskID", appHandler.GetAppCronJobTask)
		cronJobGroup.GET("/:itemID/tasks/:taskID/logs", func(ctx *gin.Context) {
			appHandler.GetAppCronJobTaskLogs(ctx, s.websocket)
		})
	}

	healthcheckGroup := appGroup.Group("/:appID/healthchecks")
	{ // healthcheck group
		healthcheckGroup.GET("", appHandler.ListAppHealthcheck)
		healthcheckGroup.GET("/:itemID", appHandler.GetAppHealthcheck)
		healthcheckGroup.POST("", appHandler.CreateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID", appHandler.UpdateAppHealthcheck)
		healthcheckGroup.PUT("/:itemID/meta", appHandler.UpdateAppHealthcheckMeta)
		healthcheckGroup.DELETE("/:itemID", appHandler.DeleteAppHealthcheck)

		// Healthcheck task group
		healthcheckGroup.GET("/:itemID/tasks", appHandler.ListAppHealthcheckTask)
	}

	appDeploymentGroup := appGroup.Group("/:appID/deployments")
	{ // app deployment group
		// Info
		appDeploymentGroup.GET("/:deploymentID", appHandler.GetAppDeployment)
		appDeploymentGroup.GET("", appHandler.ListAppDeployment)
		// Cancel
		appDeploymentGroup.POST("/:deploymentID/cancel", appHandler.CancelAppDeployment)
		// Logs
		appDeploymentGroup.GET("/:deploymentID/logs", func(ctx *gin.Context) {
			appHandler.GetAppDeploymentLogs(ctx, s.websocket)
		})
	}

	return appGroup
}
