package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerSettingRoutes(apiGroup *gin.RouterGroup) *gin.RouterGroup {
	settingGroup := apiGroup.Group("/settings")
	settingHandler := s.handlerRegistry.settingHandler

	{ // oauth group
		oauthGroup := settingGroup.Group("/oauth")
		oauthGroup.GET("/:itemID", settingHandler.GetOAuth)
		oauthGroup.GET("", settingHandler.ListOAuth)
		oauthGroup.POST("", settingHandler.CreateOAuth)
		oauthGroup.PUT("/:itemID", settingHandler.UpdateOAuth)
		oauthGroup.PUT("/:itemID/meta", settingHandler.UpdateOAuthMeta)
		oauthGroup.DELETE("/:itemID", settingHandler.DeleteOAuth)
	}

	{ // github-app group
		githubAppGroup := settingGroup.Group("/github-apps")
		githubAppGroup.GET("/:itemID", settingHandler.GetGithubApp)
		githubAppGroup.GET("", settingHandler.ListGithubApp)
		githubAppGroup.POST("", settingHandler.CreateGithubApp)
		githubAppGroup.PUT("/:itemID", settingHandler.UpdateGithubApp)
		githubAppGroup.PUT("/:itemID/meta", settingHandler.UpdateGithubAppMeta)
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

	{ // access-token group
		accessTokenGroup := settingGroup.Group("/access-tokens")
		accessTokenGroup.GET("/:itemID", settingHandler.GetAccessToken)
		accessTokenGroup.GET("", settingHandler.ListAccessToken)
		accessTokenGroup.POST("", settingHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", settingHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/meta", settingHandler.UpdateAccessTokenMeta)
		accessTokenGroup.DELETE("/:itemID", settingHandler.DeleteAccessToken)
		// Test connection
		accessTokenGroup.POST("/test-conn", settingHandler.TestAccessTokenConn)
	}

	{ // Git credentials group
		gitCredentialGroup := settingGroup.Group("/git-credentials")
		gitCredentialGroup.GET("", settingHandler.ListGitCredential)
	}

	{ // Cloud storage group
		cloudStorageGroup := settingGroup.Group("/cloud-storages")
		cloudStorageGroup.GET("/:itemID", settingHandler.GetCloudStorage)
		cloudStorageGroup.GET("", settingHandler.ListCloudStorage)
		cloudStorageGroup.POST("", settingHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", settingHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/meta", settingHandler.UpdateCloudStorageMeta)
		cloudStorageGroup.DELETE("/:itemID", settingHandler.DeleteCloudStorage)
		// Test connection
		cloudStorageGroup.POST("/test-conn", settingHandler.TestCloudStorageConn)
	}

	{ // ssh key group
		sshKeyGroup := settingGroup.Group("/ssh-keys")
		sshKeyGroup.GET("/:itemID", settingHandler.GetSSHKey)
		sshKeyGroup.GET("", settingHandler.ListSSHKey)
		sshKeyGroup.POST("", settingHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", settingHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/meta", settingHandler.UpdateSSHKeyMeta)
		sshKeyGroup.DELETE("/:itemID", settingHandler.DeleteSSHKey)
	}

	{ // IM service group
		imServiceGroup := settingGroup.Group("/im-services")
		imServiceGroup.GET("/:itemID", settingHandler.GetIMService)
		imServiceGroup.GET("", settingHandler.ListIMService)
		imServiceGroup.POST("", settingHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", settingHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/meta", settingHandler.UpdateIMServiceMeta)
		imServiceGroup.DELETE("/:itemID", settingHandler.DeleteIMService)
		// Test connection
		imServiceGroup.POST("/test-send-msg", settingHandler.TestSendInstantMsg)
	}

	{ // registry auth group
		registryAuthGroup := settingGroup.Group("/registry-auth")
		registryAuthGroup.GET("/:itemID", settingHandler.GetRegistryAuth)
		registryAuthGroup.GET("", settingHandler.ListRegistryAuth)
		registryAuthGroup.POST("", settingHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", settingHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/meta", settingHandler.UpdateRegistryAuthMeta)
		registryAuthGroup.DELETE("/:itemID", settingHandler.DeleteRegistryAuth)
		// Test connection
		registryAuthGroup.POST("/test-conn", settingHandler.TestRegistryAuthConn)
	}

	{ // basic auth group
		basicAuthGroup := settingGroup.Group("/basic-auth")
		basicAuthGroup.GET("/:itemID", settingHandler.GetBasicAuth)
		basicAuthGroup.GET("", settingHandler.ListBasicAuth)
		basicAuthGroup.POST("", settingHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", settingHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/meta", settingHandler.UpdateBasicAuthMeta)
		basicAuthGroup.DELETE("/:itemID", settingHandler.DeleteBasicAuth)
	}

	{ // ssl cert group
		sslCertGroup := settingGroup.Group("/ssl-certs")
		sslCertGroup.GET("/:itemID", settingHandler.GetSSLCert)
		sslCertGroup.GET("", settingHandler.ListSSLCert)
		sslCertGroup.POST("", settingHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", settingHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/meta", settingHandler.UpdateSSLCertMeta)
		sslCertGroup.DELETE("/:itemID", settingHandler.DeleteSSLCert)
	}

	{ // domain settings group
		domainSettingsGroup := settingGroup.Group("/domain-settings")
		domainSettingsGroup.GET("", settingHandler.GetUniqueDomainSettings)
		domainSettingsGroup.PUT("", settingHandler.UpdateUniqueDomainSettings)
		domainSettingsGroup.PUT("/meta", settingHandler.UpdateUniqueDomainSettingsMeta)
		domainSettingsGroup.DELETE("", settingHandler.DeleteUniqueDomainSettings)
	}

	{ // email group
		emailGroup := settingGroup.Group("/emails")
		emailGroup.GET("/:itemID", settingHandler.GetEmail)
		emailGroup.GET("", settingHandler.ListEmail)
		emailGroup.POST("", settingHandler.CreateEmail)
		emailGroup.PUT("/:itemID", settingHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/meta", settingHandler.UpdateEmailMeta)
		emailGroup.DELETE("/:itemID", settingHandler.DeleteEmail)
		// Test connection
		emailGroup.POST("/test-send-mail", settingHandler.TestSendMail)
	}

	{ // secrets group
		secretGroup := settingGroup.Group("/secrets")
		secretGroup.GET("", settingHandler.ListSecret)
		secretGroup.POST("", settingHandler.CreateSecret)
		secretGroup.PUT("/:itemID", settingHandler.UpdateSecret)
		secretGroup.PUT("/:itemID/meta", settingHandler.UpdateSecretMeta)
		secretGroup.DELETE("/:itemID", settingHandler.DeleteSecret)
	}

	{ // cron-job group
		cronJobGroup := settingGroup.Group("/cron-jobs")
		cronJobGroup.GET("/:itemID", settingHandler.GetCronJob)
		cronJobGroup.GET("", settingHandler.ListCronJob)
		cronJobGroup.POST("", settingHandler.CreateCronJob)
		cronJobGroup.PUT("/:itemID", settingHandler.UpdateCronJob)
		cronJobGroup.PUT("/:itemID/meta", settingHandler.UpdateCronJobMeta)
		cronJobGroup.DELETE("/:itemID", settingHandler.DeleteCronJob)
	}

	{ // notification group
		notificationGroup := settingGroup.Group("/notifications")
		notificationGroup.GET("/:itemID", settingHandler.GetNotification)
		notificationGroup.GET("", settingHandler.ListNotification)
		notificationGroup.POST("", settingHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", settingHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/meta", settingHandler.UpdateNotificationMeta)
		notificationGroup.DELETE("/:itemID", settingHandler.DeleteNotification)
	}

	{ // image-build settings group
		imageBuildSettingsGroup := settingGroup.Group("/image-build-settings")
		imageBuildSettingsGroup.GET("", settingHandler.GetUniqueImageBuildSettings)
		imageBuildSettingsGroup.PUT("", settingHandler.UpdateUniqueImageBuildSettings)
		imageBuildSettingsGroup.PUT("/meta", settingHandler.UpdateUniqueImageBuildSettingsMeta)
		imageBuildSettingsGroup.DELETE("", settingHandler.DeleteUniqueImageBuildSettings)
	}

	return settingGroup
}
