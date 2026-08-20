package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerProjectEnvRoutes(projectGroup *gin.RouterGroup) {
	projectEnvGroup := projectGroup.Group("/:projectEnv")
	projectEnvHandler := s.handlerRegistry.projectEnvHandler
	projectEnvSettingsHandler := s.handlerRegistry.projectEnvSettingsHandler

	// Project envs
	projectEnvGroup.PUT("/status", projectEnvHandler.UpdateProjectEnvStatus)
	projectEnvGroup.DELETE("", projectEnvHandler.DeleteProjectEnv)

	// Settings import
	projectEnvGroup.POST("/settings-import", projectEnvSettingsHandler.ImportSettings)

	{ // Access-token group
		accessTokenGroup := projectEnvGroup.Group("/access-tokens")
		accessTokenGroup.GET("/:itemID", projectEnvSettingsHandler.GetAccessToken)
		accessTokenGroup.GET("", projectEnvSettingsHandler.ListAccessToken)
		accessTokenGroup.POST("", projectEnvSettingsHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateAccessTokenStatus)
		accessTokenGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteAccessToken)
	}

	{ // ACME DNS Provider group
		acmeDnsProviderGroup := projectEnvGroup.Group("/acme-dns-providers")
		acmeDnsProviderGroup.GET("/:itemID", projectEnvSettingsHandler.GetAcmeDnsProvider)
		acmeDnsProviderGroup.GET("", projectEnvSettingsHandler.ListAcmeDnsProvider)
		acmeDnsProviderGroup.POST("", projectEnvSettingsHandler.CreateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateAcmeDnsProviderStatus)
		acmeDnsProviderGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteAcmeDnsProvider)
	}

	{ // Basic auth group
		basicAuthGroup := projectEnvGroup.Group("/basic-auth")
		basicAuthGroup.GET("/:itemID", projectEnvSettingsHandler.GetBasicAuth)
		basicAuthGroup.GET("", projectEnvSettingsHandler.ListBasicAuth)
		basicAuthGroup.POST("", projectEnvSettingsHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateBasicAuthStatus)
		basicAuthGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteBasicAuth)
	}

	{ // Cloud storage group
		cloudStorageGroup := projectEnvGroup.Group("/cloud-storages")
		cloudStorageGroup.GET("/:itemID", projectEnvSettingsHandler.GetCloudStorage)
		cloudStorageGroup.GET("", projectEnvSettingsHandler.ListCloudStorage)
		cloudStorageGroup.POST("", projectEnvSettingsHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateCloudStorageStatus)
		cloudStorageGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteCloudStorage)
	}

	{ // Command pipes group
		commandPipeGroup := projectEnvGroup.Group("/command-pipes")
		commandPipeGroup.GET("/:itemID", projectEnvSettingsHandler.GetCommandPipe)
		commandPipeGroup.GET("", projectEnvSettingsHandler.ListCommandPipe)
		commandPipeGroup.POST("", projectEnvSettingsHandler.CreateCommandPipe)
		commandPipeGroup.POST("/from-template", projectEnvSettingsHandler.CreateCommandPipeFromTemplate)
		commandPipeGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateCommandPipe)
		commandPipeGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateCommandPipeStatus)
		commandPipeGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteCommandPipe)
	}

	{ // Command templates group
		commandTemplateGroup := projectEnvGroup.Group("/command-templates")
		commandTemplateGroup.GET("/:itemID", projectEnvSettingsHandler.GetCommandTemplate)
		commandTemplateGroup.GET("", projectEnvSettingsHandler.ListCommandTemplate)
		commandTemplateGroup.POST("", projectEnvSettingsHandler.CreateCommandTemplate)
		commandTemplateGroup.POST("/from-template", projectEnvSettingsHandler.CreateCommandTemplateFromTemplate)
		commandTemplateGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateCommandTemplate)
		commandTemplateGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateCommandTemplateStatus)
		commandTemplateGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteCommandTemplate)
	}

	{ // Email group
		emailGroup := projectEnvGroup.Group("/emails")
		emailGroup.GET("/:itemID", projectEnvSettingsHandler.GetEmail)
		emailGroup.GET("", projectEnvSettingsHandler.ListEmail)
		emailGroup.POST("", projectEnvSettingsHandler.CreateEmail)
		emailGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateEmailStatus)
		emailGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteEmail)
	}

	{ // Env vars
		envVarGroup := projectEnvGroup.Group("/env-vars")
		envVarGroup.GET("", projectEnvSettingsHandler.GetEnvVars)
		envVarGroup.PUT("", projectEnvSettingsHandler.UpdateEnvVars)
		envVarGroup.POST("/compute", projectEnvSettingsHandler.BuildEnvVars)
	}

	{ // Git credentials group
		gitCredentialGroup := projectEnvGroup.Group("/git-credentials")
		gitCredentialGroup.GET("", projectEnvSettingsHandler.ListGitCredentials)

		// Repos
		gitCredentialGroup.GET("/:itemID/repositories", projectEnvSettingsHandler.ListGitRepository)
		// Branches
		gitCredentialGroup.GET("/:itemID/repository/branches", projectEnvSettingsHandler.ListGitBranch)
		// Pull requests
		gitCredentialGroup.GET("/:itemID/repository/pull-requests", projectEnvSettingsHandler.ListGitPullRequest)
	}

	{ // Github-app group
		githubAppGroup := projectEnvGroup.Group("/github-apps")
		githubAppGroup.GET("/:itemID", projectEnvSettingsHandler.GetGithubApp)
		githubAppGroup.GET("", projectEnvSettingsHandler.ListGithubApp)
	}

	{ // IM service group
		imServiceGroup := projectEnvGroup.Group("/im-services")
		imServiceGroup.GET("/:itemID", projectEnvSettingsHandler.GetIMService)
		imServiceGroup.GET("", projectEnvSettingsHandler.ListIMService)
		imServiceGroup.POST("", projectEnvSettingsHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateIMServiceStatus)
		imServiceGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteIMService)
	}

	{ // Notification group
		notificationGroup := projectEnvGroup.Group("/notifications")
		notificationGroup.GET("/:itemID", projectEnvSettingsHandler.GetNotification)
		notificationGroup.GET("", projectEnvSettingsHandler.ListNotification)
		notificationGroup.POST("", projectEnvSettingsHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateNotificationStatus)
		notificationGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteNotification)
	}

	{ // Registry auth group
		registryAuthGroup := projectEnvGroup.Group("/registry-auth")
		registryAuthGroup.GET("/:itemID", projectEnvSettingsHandler.GetRegistryAuth)
		registryAuthGroup.GET("", projectEnvSettingsHandler.ListRegistryAuth)
		registryAuthGroup.POST("", projectEnvSettingsHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateRegistryAuthStatus)
		registryAuthGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteRegistryAuth)
	}

	{ // Repo webhook group
		repoWebhookGroup := projectEnvGroup.Group("/repo-webhooks")
		repoWebhookGroup.GET("/:itemID", projectEnvSettingsHandler.GetRepoWebhook)
		repoWebhookGroup.GET("", projectEnvSettingsHandler.ListRepoWebhook)
	}

	{ // Secrets
		secretGroup := projectEnvGroup.Group("/secrets")
		secretGroup.GET("", projectEnvSettingsHandler.ListSecret)
		secretGroup.GET("/:itemID", projectEnvSettingsHandler.GetSecret)
		secretGroup.POST("", projectEnvSettingsHandler.CreateSecret)
		secretGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSecret)
		secretGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSecretStatus)
		secretGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSecret)
	}

	{ // SSH key group
		sshKeyGroup := projectEnvGroup.Group("/ssh-keys")
		sshKeyGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSHKey)
		sshKeyGroup.GET("", projectEnvSettingsHandler.ListSSHKey)
		sshKeyGroup.POST("", projectEnvSettingsHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSHKeyStatus)
		sshKeyGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSHKey)
	}

	{ // SSL Provider group
		sslProviderGroup := projectEnvGroup.Group("/ssl-providers")
		sslProviderGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSLProvider)
		sslProviderGroup.GET("", projectEnvSettingsHandler.ListSSLProvider)
		sslProviderGroup.POST("", projectEnvSettingsHandler.CreateSSLProvider)
		sslProviderGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSLProvider)
		sslProviderGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSLProviderStatus)
		sslProviderGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSLProvider)
	}

	{ // SSL Cert group
		sslCertGroup := projectEnvGroup.Group("/ssl-certs")
		sslCertGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSLCert)
		sslCertGroup.GET("", projectEnvSettingsHandler.ListSSLCert)
		sslCertGroup.POST("", projectEnvSettingsHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSLCertStatus)
		sslCertGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSLCert)
		sslCertGroup.POST("/:itemID/renew", projectEnvSettingsHandler.RenewSSLCert)
	}

	_ = s.registerAppRoutes(projectGroup, projectEnvGroup)
}
