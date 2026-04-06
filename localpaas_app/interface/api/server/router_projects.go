package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerProjectRoutes(apiGroup *gin.RouterGroup) *gin.RouterGroup {
	projectGroup := apiGroup.Group("/projects")
	projectHandler := s.handlerRegistry.projectHandler

	// Projects
	projectGroup.GET("/base", projectHandler.ListProjectBase)
	projectGroup.GET("/:projectID", projectHandler.GetProject)
	projectGroup.GET("", projectHandler.ListProject)
	projectGroup.POST("", projectHandler.CreateProject)
	projectGroup.PUT("/:projectID", projectHandler.UpdateProject)
	projectGroup.PUT("/:projectID/photo", projectHandler.UpdateProjectPhoto)
	projectGroup.DELETE("/:projectID", projectHandler.DeleteProject)

	// Settings import
	projectGroup.POST("/:projectID/settings-import", projectHandler.ImportSettings)

	{ // Tags
		tagGroup := projectGroup.Group("/:projectID/tags")
		tagGroup.POST("", projectHandler.CreateProjectTag)
		tagGroup.POST("/delete", projectHandler.DeleteProjectTags)
	}

	{ // Env vars
		envVarGroup := projectGroup.Group("/:projectID/env-vars")
		envVarGroup.GET("", projectHandler.GetProjectEnvVars)
		envVarGroup.PUT("", projectHandler.UpdateProjectEnvVars)
	}

	{ // Secrets
		secretGroup := projectGroup.Group("/:projectID/secrets")
		secretGroup.GET("", projectHandler.ListProjectSecrets)
		secretGroup.POST("", projectHandler.CreateProjectSecret)
		secretGroup.PUT("/:itemID", projectHandler.UpdateProjectSecret)
		secretGroup.DELETE("/:itemID", projectHandler.DeleteProjectSecret)
	}

	{ // Cron jobs
		cronJobGroup := projectGroup.Group("/:projectID/cron-jobs")
		cronJobGroup.GET("", projectHandler.ListCronJob)
		cronJobGroup.GET("/:itemID", projectHandler.GetCronJob)
		cronJobGroup.POST("", projectHandler.CreateCronJob)
		cronJobGroup.PUT("/:itemID", projectHandler.UpdateCronJob)
		cronJobGroup.PUT("/:itemID/meta", projectHandler.UpdateCronJobMeta)
		cronJobGroup.DELETE("/:itemID", projectHandler.DeleteCronJob)
	}

	{ // Github-app group
		githubAppGroup := projectGroup.Group("/:projectID/github-apps")
		githubAppGroup.GET("/:itemID", projectHandler.GetGithubApp)
		githubAppGroup.GET("", projectHandler.ListGithubApp)
		githubAppGroup.POST("", projectHandler.CreateGithubApp)
		githubAppGroup.PUT("/:itemID", projectHandler.UpdateGithubApp)
		githubAppGroup.PUT("/:itemID/meta", projectHandler.UpdateGithubAppMeta)
		githubAppGroup.DELETE("/:itemID", projectHandler.DeleteGithubApp)
		// Manifest flow
		githubAppGroup.POST("/manifest-flow/begin", projectHandler.BeginProjectGithubAppManifestFlow)
		githubAppGroup.GET("/:itemID/manifest-flow/begin", projectHandler.BeginProjectGithubAppManifestFlowCreation)
		githubAppGroup.GET("/:itemID/manifest-flow/progress", projectHandler.HandleProjectGithubAppManifestFlowProgress)
		githubAppGroup.POST("/:itemID/begin-reprovision", projectHandler.BeginReprovisionProjectGithubApp)
	}

	{ // Access-token group
		accessTokenGroup := projectGroup.Group("/:projectID/access-tokens")
		accessTokenGroup.GET("/:itemID", projectHandler.GetAccessToken)
		accessTokenGroup.GET("", projectHandler.ListAccessToken)
		accessTokenGroup.POST("", projectHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", projectHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/meta", projectHandler.UpdateAccessTokenMeta)
		accessTokenGroup.DELETE("/:itemID", projectHandler.DeleteAccessToken)
	}

	{ // Git credentials group
		gitCredentialGroup := projectGroup.Group("/:projectID/git-credentials")
		gitCredentialGroup.GET("", projectHandler.ListGitCredential)

		// Repos
		gitCredentialGroup.GET("/:itemID/repositories", projectHandler.ListGitRepo)
	}

	{ // Cloud storage group
		cloudStorageGroup := projectGroup.Group("/:projectID/cloud-storages")
		cloudStorageGroup.GET("/:itemID", projectHandler.GetCloudStorage)
		cloudStorageGroup.GET("", projectHandler.ListCloudStorage)
		cloudStorageGroup.POST("", projectHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", projectHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/meta", projectHandler.UpdateCloudStorageMeta)
		cloudStorageGroup.DELETE("/:itemID", projectHandler.DeleteCloudStorage)
	}

	{ // SSH key group
		sshKeyGroup := projectGroup.Group("/:projectID/ssh-keys")
		sshKeyGroup.GET("/:itemID", projectHandler.GetSSHKey)
		sshKeyGroup.GET("", projectHandler.ListSSHKey)
		sshKeyGroup.POST("", projectHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", projectHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/meta", projectHandler.UpdateSSHKeyMeta)
		sshKeyGroup.DELETE("/:itemID", projectHandler.DeleteSSHKey)
	}

	{ // IM service group
		imServiceGroup := projectGroup.Group("/:projectID/im-services")
		imServiceGroup.GET("/:itemID", projectHandler.GetIMService)
		imServiceGroup.GET("", projectHandler.ListIMService)
		imServiceGroup.POST("", projectHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", projectHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/meta", projectHandler.UpdateIMServiceMeta)
		imServiceGroup.DELETE("/:itemID", projectHandler.DeleteIMService)
	}

	{ // Registry auth group
		registryAuthGroup := projectGroup.Group("/:projectID/registry-auth")
		registryAuthGroup.GET("/:itemID", projectHandler.GetRegistryAuth)
		registryAuthGroup.GET("", projectHandler.ListRegistryAuth)
		registryAuthGroup.POST("", projectHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", projectHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/meta", projectHandler.UpdateRegistryAuthMeta)
		registryAuthGroup.DELETE("/:itemID", projectHandler.DeleteRegistryAuth)
	}

	{ // Basic auth group
		basicAuthGroup := projectGroup.Group("/:projectID/basic-auth")
		basicAuthGroup.GET("/:itemID", projectHandler.GetBasicAuth)
		basicAuthGroup.GET("", projectHandler.ListBasicAuth)
		basicAuthGroup.POST("", projectHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", projectHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/meta", projectHandler.UpdateBasicAuthMeta)
		basicAuthGroup.DELETE("/:itemID", projectHandler.DeleteBasicAuth)
	}

	{ // SSL Cert group
		sslCertGroup := projectGroup.Group("/:projectID/ssl-certs")
		sslCertGroup.GET("/:itemID", projectHandler.GetSSLCert)
		sslCertGroup.GET("", projectHandler.ListSSLCert)
		sslCertGroup.POST("", projectHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", projectHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/meta", projectHandler.UpdateSSLCertMeta)
		sslCertGroup.DELETE("/:itemID", projectHandler.DeleteSSLCert)
	}

	{ // Email group
		emailGroup := projectGroup.Group("/:projectID/emails")
		emailGroup.GET("/:itemID", projectHandler.GetEmail)
		emailGroup.GET("", projectHandler.ListEmail)
		emailGroup.POST("", projectHandler.CreateEmail)
		emailGroup.PUT("/:itemID", projectHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/meta", projectHandler.UpdateEmailMeta)
		emailGroup.DELETE("/:itemID", projectHandler.DeleteEmail)
	}

	{ // Repo webhook group
		repoWebhookGroup := projectGroup.Group("/:projectID/repo-webhooks")
		repoWebhookGroup.GET("/:itemID", projectHandler.GetRepoWebhook)
		repoWebhookGroup.GET("", projectHandler.ListRepoWebhook)
		repoWebhookGroup.POST("", projectHandler.CreateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID", projectHandler.UpdateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID/meta", projectHandler.UpdateRepoWebhookMeta)
		repoWebhookGroup.DELETE("/:itemID", projectHandler.DeleteRepoWebhook)
	}

	{ // Notification group
		notificationGroup := projectGroup.Group("/:projectID/notifications")
		notificationGroup.GET("/:itemID", projectHandler.GetNotification)
		notificationGroup.GET("", projectHandler.ListNotification)
		notificationGroup.POST("", projectHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", projectHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/meta", projectHandler.UpdateNotificationMeta)
		notificationGroup.DELETE("/:itemID", projectHandler.DeleteNotification)
	}

	{ // Image build group
		imageBuildGroup := projectGroup.Group("/:projectID/image-build")
		imageBuildGroup.GET("/:itemID", projectHandler.GetImageBuild)
		imageBuildGroup.GET("", projectHandler.ListImageBuild)
		imageBuildGroup.POST("", projectHandler.CreateImageBuild)
		imageBuildGroup.PUT("/:itemID", projectHandler.UpdateImageBuild)
		imageBuildGroup.PUT("/:itemID/meta", projectHandler.UpdateImageBuildMeta)
		imageBuildGroup.DELETE("/:itemID", projectHandler.DeleteImageBuild)
	}

	{ // Network group
		networkGroup := projectGroup.Group("/:projectID/networks")
		networkGroup.GET("/:networkID", projectHandler.GetNetwork)
		networkGroup.GET("/:networkID/inspect", projectHandler.GetNetworkInspection)
		networkGroup.GET("", projectHandler.ListNetwork)
		networkGroup.POST("", projectHandler.CreateNetwork)
		networkGroup.DELETE("/:networkID", projectHandler.DeleteNetwork)
	}

	{ // Volume group
		volumeGroup := projectGroup.Group("/:projectID/volumes")
		volumeGroup.GET("/:volumeID", projectHandler.GetVolume)
		volumeGroup.GET("/:volumeID/inspect", projectHandler.GetVolumeInspection)
		volumeGroup.GET("", projectHandler.ListVolume)
		volumeGroup.POST("", projectHandler.CreateVolume)
		volumeGroup.DELETE("/:volumeID", projectHandler.DeleteVolume)
	}

	return projectGroup
}
