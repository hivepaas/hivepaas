package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerProjectRoutes(apiGroup *gin.RouterGroup) {
	projectGroup := apiGroup.Group("/projects")
	projectHandler := s.handlerRegistry.projectHandler
	projectSettingsHandler := s.handlerRegistry.projectSettingsHandler

	// Projects
	projectGroup.GET("/base", projectHandler.ListProjectBase)
	projectGroup.GET("/:projectID", projectHandler.GetProject)
	projectGroup.GET("", projectHandler.ListProject)
	projectGroup.POST("", projectHandler.CreateProject)
	projectGroup.PUT("/:projectID", projectHandler.UpdateProject)
	// Status Update
	projectGroup.PUT("/:projectID/status", projectHandler.UpdateProjectStatus)
	// Photo Update
	projectGroup.PUT("/:projectID/photo", projectHandler.UpdateProjectPhoto)
	projectGroup.DELETE("/:projectID", projectHandler.DeleteProject)

	// Settings import
	projectGroup.POST("/:projectID/settings-import", projectSettingsHandler.ImportSettings)

	{ // Tags
		tagGroup := projectGroup.Group("/:projectID/tags")
		tagGroup.POST("", projectSettingsHandler.CreateProjectTag)
		tagGroup.POST("/delete", projectSettingsHandler.DeleteProjectTags)
	}

	{ // User accesses
		userAccessGroup := projectGroup.Group("/:projectID/user-accesses")
		userAccessGroup.GET("", projectSettingsHandler.GetProjectUserAccesses)
		userAccessGroup.PUT("", projectSettingsHandler.UpdateProjectUserAccesses)
	}

	{ // Env vars
		envVarGroup := projectGroup.Group("/:projectID/env-vars")
		envVarGroup.GET("", projectSettingsHandler.GetEnvVars)
		envVarGroup.PUT("", projectSettingsHandler.UpdateEnvVars)
	}

	{ // Secrets
		secretGroup := projectGroup.Group("/:projectID/secrets")
		secretGroup.GET("", projectSettingsHandler.ListSecret)
		secretGroup.GET("/:itemID", projectSettingsHandler.GetSecret)
		secretGroup.POST("", projectSettingsHandler.CreateSecret)
		secretGroup.PUT("/:itemID", projectSettingsHandler.UpdateSecret)
		secretGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateSecretStatus)
		secretGroup.DELETE("/:itemID", projectSettingsHandler.DeleteSecret)
	}

	{ // Scheduled jobs
		schedJobGroup := projectGroup.Group("/:projectID/sched-jobs")
		schedJobGroup.GET("", projectSettingsHandler.ListSchedJob)
		schedJobGroup.GET("/:itemID", projectSettingsHandler.GetSchedJob)
		schedJobGroup.POST("", projectSettingsHandler.CreateSchedJob)
		schedJobGroup.PUT("/:itemID", projectSettingsHandler.UpdateSchedJob)
		schedJobGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateSchedJobStatus)
		schedJobGroup.DELETE("/:itemID", projectSettingsHandler.DeleteSchedJob)
	}

	{ // Github-app group
		githubAppGroup := projectGroup.Group("/:projectID/github-apps")
		githubAppGroup.GET("/:itemID", projectSettingsHandler.GetGithubApp)
		githubAppGroup.GET("", projectSettingsHandler.ListGithubApp)
		githubAppGroup.POST("", projectSettingsHandler.CreateGithubApp)
		githubAppGroup.PUT("/:itemID", projectSettingsHandler.UpdateGithubApp)
		githubAppGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateGithubAppStatus)
		githubAppGroup.DELETE("/:itemID", projectSettingsHandler.DeleteGithubApp)
		// Manifest flow
		githubAppGroup.POST("/manifest-flow/begin",
			projectSettingsHandler.BeginProjectGithubAppManifestFlow)
		githubAppGroup.GET("/:itemID/manifest-flow/begin",
			projectSettingsHandler.BeginProjectGithubAppManifestFlowCreation)
		githubAppGroup.GET("/:itemID/manifest-flow/progress",
			projectSettingsHandler.HandleProjectGithubAppManifestFlowProgress)
		githubAppGroup.POST("/:itemID/begin-reprovision",
			projectSettingsHandler.BeginReprovisionProjectGithubApp)
	}

	{ // Access-token group
		accessTokenGroup := projectGroup.Group("/:projectID/access-tokens")
		accessTokenGroup.GET("/:itemID", projectSettingsHandler.GetAccessToken)
		accessTokenGroup.GET("", projectSettingsHandler.ListAccessToken)
		accessTokenGroup.POST("", projectSettingsHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", projectSettingsHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateAccessTokenStatus)
		accessTokenGroup.DELETE("/:itemID", projectSettingsHandler.DeleteAccessToken)
	}

	{ // Git credentials group
		gitCredentialGroup := projectGroup.Group("/:projectID/git-credentials")
		gitCredentialGroup.GET("", projectSettingsHandler.ListGitCredentials)

		// Repos
		gitCredentialGroup.GET("/:itemID/repositories", projectSettingsHandler.ListGitRepo)
	}

	{ // Cloud storage group
		cloudStorageGroup := projectGroup.Group("/:projectID/cloud-storages")
		cloudStorageGroup.GET("/:itemID", projectSettingsHandler.GetCloudStorage)
		cloudStorageGroup.GET("", projectSettingsHandler.ListCloudStorage)
		cloudStorageGroup.POST("", projectSettingsHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", projectSettingsHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateCloudStorageStatus)
		cloudStorageGroup.DELETE("/:itemID", projectSettingsHandler.DeleteCloudStorage)
	}

	{ // SSH key group
		sshKeyGroup := projectGroup.Group("/:projectID/ssh-keys")
		sshKeyGroup.GET("/:itemID", projectSettingsHandler.GetSSHKey)
		sshKeyGroup.GET("", projectSettingsHandler.ListSSHKey)
		sshKeyGroup.POST("", projectSettingsHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", projectSettingsHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateSSHKeyStatus)
		sshKeyGroup.DELETE("/:itemID", projectSettingsHandler.DeleteSSHKey)
	}

	{ // IM service group
		imServiceGroup := projectGroup.Group("/:projectID/im-services")
		imServiceGroup.GET("/:itemID", projectSettingsHandler.GetIMService)
		imServiceGroup.GET("", projectSettingsHandler.ListIMService)
		imServiceGroup.POST("", projectSettingsHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", projectSettingsHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateIMServiceStatus)
		imServiceGroup.DELETE("/:itemID", projectSettingsHandler.DeleteIMService)
	}

	{ // Registry auth group
		registryAuthGroup := projectGroup.Group("/:projectID/registry-auth")
		registryAuthGroup.GET("/:itemID", projectSettingsHandler.GetRegistryAuth)
		registryAuthGroup.GET("", projectSettingsHandler.ListRegistryAuth)
		registryAuthGroup.POST("", projectSettingsHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", projectSettingsHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateRegistryAuthStatus)
		registryAuthGroup.DELETE("/:itemID", projectSettingsHandler.DeleteRegistryAuth)
	}

	{ // Basic auth group
		basicAuthGroup := projectGroup.Group("/:projectID/basic-auth")
		basicAuthGroup.GET("/:itemID", projectSettingsHandler.GetBasicAuth)
		basicAuthGroup.GET("", projectSettingsHandler.ListBasicAuth)
		basicAuthGroup.POST("", projectSettingsHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", projectSettingsHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateBasicAuthStatus)
		basicAuthGroup.DELETE("/:itemID", projectSettingsHandler.DeleteBasicAuth)
	}

	{ // SSL Cert group
		sslCertGroup := projectGroup.Group("/:projectID/ssl-certs")
		sslCertGroup.GET("/:itemID", projectSettingsHandler.GetSSLCert)
		sslCertGroup.GET("", projectSettingsHandler.ListSSLCert)
		sslCertGroup.POST("", projectSettingsHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", projectSettingsHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateSSLCertStatus)
		sslCertGroup.DELETE("/:itemID", projectSettingsHandler.DeleteSSLCert)
	}

	{ // Domain settings group
		domainSettingsGroup := projectGroup.Group("/:projectID/domain-settings")
		domainSettingsGroup.GET("", projectSettingsHandler.GetDomainSettings)
		domainSettingsGroup.PUT("", projectSettingsHandler.UpdateDomainSettings)
		domainSettingsGroup.PUT("/status", projectSettingsHandler.UpdateDomainSettingsStatus)
		domainSettingsGroup.DELETE("", projectSettingsHandler.DeleteDomainSettings)
	}

	{ // Storage settings group
		storageSettingsGroup := projectGroup.Group("/:projectID/storage-settings")
		storageSettingsGroup.GET("", projectSettingsHandler.GetStorageSettings)
		storageSettingsGroup.PUT("", projectSettingsHandler.UpdateStorageSettings)
		storageSettingsGroup.PUT("/status", projectSettingsHandler.UpdateStorageSettingsStatus)
		storageSettingsGroup.DELETE("", projectSettingsHandler.DeleteStorageSettings)
	}

	{ // Email group
		emailGroup := projectGroup.Group("/:projectID/emails")
		emailGroup.GET("/:itemID", projectSettingsHandler.GetEmail)
		emailGroup.GET("", projectSettingsHandler.ListEmail)
		emailGroup.POST("", projectSettingsHandler.CreateEmail)
		emailGroup.PUT("/:itemID", projectSettingsHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateEmailStatus)
		emailGroup.DELETE("/:itemID", projectSettingsHandler.DeleteEmail)
	}

	{ // Repo webhook group
		repoWebhookGroup := projectGroup.Group("/:projectID/repo-webhooks")
		repoWebhookGroup.GET("/:itemID", projectSettingsHandler.GetRepoWebhook)
		repoWebhookGroup.GET("", projectSettingsHandler.ListRepoWebhook)
		repoWebhookGroup.POST("", projectSettingsHandler.CreateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID", projectSettingsHandler.UpdateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateRepoWebhookStatus)
		repoWebhookGroup.DELETE("/:itemID", projectSettingsHandler.DeleteRepoWebhook)
	}

	{ // Notification group
		notificationGroup := projectGroup.Group("/:projectID/notifications")
		notificationGroup.GET("/:itemID", projectSettingsHandler.GetNotification)
		notificationGroup.GET("", projectSettingsHandler.ListNotification)
		notificationGroup.POST("", projectSettingsHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", projectSettingsHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/status", projectSettingsHandler.UpdateNotificationStatus)
		notificationGroup.DELETE("/:itemID", projectSettingsHandler.DeleteNotification)
	}

	{ // Image build settings group
		imageBuildSettingsGroup := projectGroup.Group("/:projectID/image-build-settings")
		imageBuildSettingsGroup.GET("", projectSettingsHandler.GetImageBuildSettings)
		imageBuildSettingsGroup.PUT("", projectSettingsHandler.UpdateImageBuildSettings)
		imageBuildSettingsGroup.PUT("/status", projectSettingsHandler.UpdateImageBuildSettingsStatus)
		imageBuildSettingsGroup.DELETE("", projectSettingsHandler.DeleteImageBuildSettings)
		// Repo cache
		imageBuildSettingsGroup.GET("/repo-cache", projectSettingsHandler.GetRepoCacheInfo)
		imageBuildSettingsGroup.POST("/repo-cache/clear", projectSettingsHandler.ClearRepoCache)
	}

	{ // Docker network group
		networkGroup := projectGroup.Group("/:projectID/docker-networks")
		networkGroup.GET("/:networkID", projectSettingsHandler.GetDockerNetwork)
		networkGroup.GET("/:networkID/inspect", projectSettingsHandler.GetDockerNetworkInspection)
		networkGroup.GET("", projectSettingsHandler.ListDockerNetwork)
		networkGroup.POST("", projectSettingsHandler.CreateDockerNetwork)
		networkGroup.DELETE("/:networkID", projectSettingsHandler.DeleteDockerNetwork)
	}

	{ // Docker volume group
		volumeGroup := projectGroup.Group("/:projectID/docker-volumes")
		volumeGroup.GET("/:volumeID", projectSettingsHandler.GetDockerVolume)
		volumeGroup.GET("/:volumeID/inspect", projectSettingsHandler.GetDockerVolumeInspection)
		volumeGroup.GET("", projectSettingsHandler.ListDockerVolume)
		volumeGroup.POST("", projectSettingsHandler.CreateDockerVolume)
		volumeGroup.DELETE("/:volumeID", projectSettingsHandler.DeleteDockerVolume)
	}

	_ = s.registerAppRoutes(projectGroup)
}
