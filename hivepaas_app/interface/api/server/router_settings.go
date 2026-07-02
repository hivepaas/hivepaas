package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerSettingRoutes(apiGroup *gin.RouterGroup) {
	settingGroup := apiGroup.Group("/settings")
	settingHandler := s.handlerRegistry.settingHandler

	{ // accessing projects group
		settingGroup.GET("/:itemID/accessible-by-projects", settingHandler.GetAccessibleByProjects)
		settingGroup.PUT("/:itemID/accessible-by-projects", settingHandler.UpdateAccessibleByProjects)
	}

	{ // access-token group
		accessTokenGroup := settingGroup.Group("/access-tokens")
		accessTokenGroup.GET("/:itemID", settingHandler.GetAccessToken)
		accessTokenGroup.GET("", settingHandler.ListAccessToken)
		accessTokenGroup.POST("", settingHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", settingHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/status", settingHandler.UpdateAccessTokenStatus)
		accessTokenGroup.DELETE("/:itemID", settingHandler.DeleteAccessToken)
		// Test connection
		accessTokenGroup.POST("/test-conn", settingHandler.TestAccessTokenConn)
	}

	{ // acme dns provider group
		acmeDnsProviderGroup := settingGroup.Group("/acme-dns-providers")
		acmeDnsProviderGroup.GET("/:itemID", settingHandler.GetAcmeDnsProvider)
		acmeDnsProviderGroup.GET("", settingHandler.ListAcmeDnsProvider)
		acmeDnsProviderGroup.POST("", settingHandler.CreateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID", settingHandler.UpdateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID/status", settingHandler.UpdateAcmeDnsProviderStatus)
		acmeDnsProviderGroup.DELETE("/:itemID", settingHandler.DeleteAcmeDnsProvider)
		// Test access
		acmeDnsProviderGroup.POST("/test-access", settingHandler.TestAcmeDnsProviderAccess)
	}

	{ // basic auth group
		basicAuthGroup := settingGroup.Group("/basic-auth")
		basicAuthGroup.GET("/:itemID", settingHandler.GetBasicAuth)
		basicAuthGroup.GET("", settingHandler.ListBasicAuth)
		basicAuthGroup.POST("", settingHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", settingHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/status", settingHandler.UpdateBasicAuthStatus)
		basicAuthGroup.DELETE("/:itemID", settingHandler.DeleteBasicAuth)
	}

	{ // Cloud storage group
		cloudStorageGroup := settingGroup.Group("/cloud-storages")
		cloudStorageGroup.GET("/:itemID", settingHandler.GetCloudStorage)
		cloudStorageGroup.GET("", settingHandler.ListCloudStorage)
		cloudStorageGroup.POST("", settingHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", settingHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/status", settingHandler.UpdateCloudStorageStatus)
		cloudStorageGroup.DELETE("/:itemID", settingHandler.DeleteCloudStorage)
		// Test connection
		cloudStorageGroup.POST("/test-conn", settingHandler.TestCloudStorageConn)
	}

	{ // domain settings group
		domainSettingsGroup := settingGroup.Group("/domain-settings")
		domainSettingsGroup.GET("", settingHandler.GetDomainSettings)
		domainSettingsGroup.PUT("", settingHandler.UpdateDomainSettings)
		domainSettingsGroup.PUT("/status", settingHandler.UpdateDomainSettingsStatus)
		domainSettingsGroup.DELETE("", settingHandler.DeleteDomainSettings)
	}

	{ // email group
		emailGroup := settingGroup.Group("/emails")
		emailGroup.GET("/:itemID", settingHandler.GetEmail)
		emailGroup.GET("", settingHandler.ListEmail)
		emailGroup.POST("", settingHandler.CreateEmail)
		emailGroup.PUT("/:itemID", settingHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/status", settingHandler.UpdateEmailStatus)
		emailGroup.DELETE("/:itemID", settingHandler.DeleteEmail)
		// Test connection
		emailGroup.POST("/test-send-mail", settingHandler.TestSendMail)
	}

	{ // Git credentials group
		gitCredentialGroup := settingGroup.Group("/git-credentials")
		gitCredentialGroup.GET("", settingHandler.ListGitCredential)
	}

	{ // github-app group
		githubAppGroup := settingGroup.Group("/github-apps")
		githubAppGroup.GET("/:itemID", settingHandler.GetGithubApp)
		githubAppGroup.GET("", settingHandler.ListGithubApp)
		githubAppGroup.POST("", settingHandler.CreateGithubApp)
		githubAppGroup.PUT("/:itemID", settingHandler.UpdateGithubApp)
		githubAppGroup.PUT("/:itemID/status", settingHandler.UpdateGithubAppStatus)
		githubAppGroup.DELETE("/:itemID", settingHandler.DeleteGithubApp)
		// Test connection
		githubAppGroup.POST("/test-conn", settingHandler.TestGithubAppConn)
		// Installation
		githubAppGroup.POST("/installations/list", settingHandler.ListAppInstallation)
		// Manifest flow
		githubAppGroup.POST("/manifest-flow/begin", settingHandler.BeginGithubAppManifestFlow)
		githubAppGroup.GET("/:itemID/manifest-flow/begin", settingHandler.BeginGithubAppManifestFlowCreation)
		githubAppGroup.GET("/:itemID/manifest-flow/progress", settingHandler.HandleGithubAppManifestFlowProgress)
		githubAppGroup.POST("/:itemID/begin-reprovision", settingHandler.BeginReprovisionGithubApp)
	}

	{ // IM service group
		imServiceGroup := settingGroup.Group("/im-services")
		imServiceGroup.GET("/:itemID", settingHandler.GetIMService)
		imServiceGroup.GET("", settingHandler.ListIMService)
		imServiceGroup.POST("", settingHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", settingHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/status", settingHandler.UpdateIMServiceStatus)
		imServiceGroup.DELETE("/:itemID", settingHandler.DeleteIMService)
		// Test connection
		imServiceGroup.POST("/test-send-msg", settingHandler.TestSendInstantMsg)
	}

	{ // image-build settings group
		imageBuildSettingsGroup := settingGroup.Group("/image-build-settings")
		imageBuildSettingsGroup.GET("", settingHandler.GetImageBuildSettings)
		imageBuildSettingsGroup.PUT("", settingHandler.UpdateImageBuildSettings)
		imageBuildSettingsGroup.PUT("/status", settingHandler.UpdateImageBuildSettingsStatus)
		imageBuildSettingsGroup.DELETE("", settingHandler.DeleteImageBuildSettings)
		// Repo cache
		imageBuildSettingsGroup.GET("/repo-cache", settingHandler.GetRepoCacheInfo)
		imageBuildSettingsGroup.POST("/repo-cache/clear", settingHandler.ClearRepoCache)
	}

	{ // oauth group
		oauthGroup := settingGroup.Group("/oauth")
		oauthGroup.GET("/:itemID", settingHandler.GetOAuth)
		oauthGroup.GET("", settingHandler.ListOAuth)
		oauthGroup.POST("", settingHandler.CreateOAuth)
		oauthGroup.PUT("/:itemID", settingHandler.UpdateOAuth)
		oauthGroup.PUT("/:itemID/status", settingHandler.UpdateOAuthStatus)
		oauthGroup.DELETE("/:itemID", settingHandler.DeleteOAuth)
	}

	{ // notification group
		notificationGroup := settingGroup.Group("/notifications")
		notificationGroup.GET("/:itemID", settingHandler.GetNotification)
		notificationGroup.GET("", settingHandler.ListNotification)
		notificationGroup.POST("", settingHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", settingHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/status", settingHandler.UpdateNotificationStatus)
		notificationGroup.DELETE("/:itemID", settingHandler.DeleteNotification)
	}

	{ // registry auth group
		registryAuthGroup := settingGroup.Group("/registry-auth")
		registryAuthGroup.GET("/:itemID", settingHandler.GetRegistryAuth)
		registryAuthGroup.GET("", settingHandler.ListRegistryAuth)
		registryAuthGroup.POST("", settingHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", settingHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/status", settingHandler.UpdateRegistryAuthStatus)
		registryAuthGroup.DELETE("/:itemID", settingHandler.DeleteRegistryAuth)
		// Test connection
		registryAuthGroup.POST("/test-conn", settingHandler.TestRegistryAuthConn)
	}

	{ // repo webhook group
		repoWebhookGroup := settingGroup.Group("/repo-webhooks")
		repoWebhookGroup.GET("/:itemID", settingHandler.GetRepoWebhook)
		repoWebhookGroup.GET("", settingHandler.ListRepoWebhook)
		repoWebhookGroup.POST("", settingHandler.CreateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID", settingHandler.UpdateRepoWebhook)
		repoWebhookGroup.PUT("/:itemID/status", settingHandler.UpdateRepoWebhookStatus)
		repoWebhookGroup.DELETE("/:itemID", settingHandler.DeleteRepoWebhook)
	}

	{ // sched-job group
		schedJobGroup := settingGroup.Group("/sched-jobs")
		schedJobGroup.GET("/:itemID", settingHandler.GetSchedJob)
		schedJobGroup.GET("", settingHandler.ListSchedJob)
		schedJobGroup.POST("", settingHandler.CreateSchedJob)
		schedJobGroup.PUT("/:itemID", settingHandler.UpdateSchedJob)
		schedJobGroup.PUT("/:itemID/status", settingHandler.UpdateSchedJobStatus)
		schedJobGroup.DELETE("/:itemID", settingHandler.DeleteSchedJob)
		schedJobGroup.POST("calc-next-runs", settingHandler.SchedJobCalcNextRuns)
	}

	{ // secrets group
		secretGroup := settingGroup.Group("/secrets")
		secretGroup.GET("", settingHandler.ListSecret)
		secretGroup.POST("", settingHandler.CreateSecret)
		secretGroup.PUT("/:itemID", settingHandler.UpdateSecret)
		secretGroup.PUT("/:itemID/status", settingHandler.UpdateSecretStatus)
		secretGroup.DELETE("/:itemID", settingHandler.DeleteSecret)
	}

	{ // SSH key group
		sshKeyGroup := settingGroup.Group("/ssh-keys")
		sshKeyGroup.GET("/:itemID", settingHandler.GetSSHKey)
		sshKeyGroup.GET("", settingHandler.ListSSHKey)
		sshKeyGroup.POST("", settingHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", settingHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/status", settingHandler.UpdateSSHKeyStatus)
		sshKeyGroup.DELETE("/:itemID", settingHandler.DeleteSSHKey)
		sshKeyGroup.POST("/generate", settingHandler.GenerateSSHKey)
	}

	{ // ssl cert group
		sslCertGroup := settingGroup.Group("/ssl-certs")
		sslCertGroup.GET("/:itemID", settingHandler.GetSSLCert)
		sslCertGroup.GET("", settingHandler.ListSSLCert)
		sslCertGroup.POST("", settingHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", settingHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/status", settingHandler.UpdateSSLCertStatus)
		sslCertGroup.DELETE("/:itemID", settingHandler.DeleteSSLCert)
	}

	{ // ssl provider group
		sslProviderGroup := settingGroup.Group("/ssl-providers")
		sslProviderGroup.GET("/:itemID", settingHandler.GetSSLProvider)
		sslProviderGroup.GET("", settingHandler.ListSSLProvider)
		sslProviderGroup.POST("", settingHandler.CreateSSLProvider)
		sslProviderGroup.PUT("/:itemID", settingHandler.UpdateSSLProvider)
		sslProviderGroup.PUT("/:itemID/status", settingHandler.UpdateSSLProviderStatus)
		sslProviderGroup.DELETE("/:itemID", settingHandler.DeleteSSLProvider)
	}

	{ // storage settings group
		storageSettingsGroup := settingGroup.Group("/storage-settings")
		storageSettingsGroup.GET("", settingHandler.GetStorageSettings)
		storageSettingsGroup.PUT("", settingHandler.UpdateStorageSettings)
		storageSettingsGroup.PUT("/status", settingHandler.UpdateStorageSettingsStatus)
		storageSettingsGroup.DELETE("", settingHandler.DeleteStorageSettings)
	}
}
