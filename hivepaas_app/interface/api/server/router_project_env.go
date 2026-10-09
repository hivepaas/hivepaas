package server

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

	// Configuration spec export
	projectEnvGroup.POST("/spec/export", s.handlerRegistry.specHandler.ExportProjectEnvSpec)
	projectEnvGroup.POST("/spec/import/validate", s.handlerRegistry.specHandler.ValidateProjectEnvImport)
	projectEnvGroup.POST("/spec/import/apply", s.handlerRegistry.specHandler.ApplyProjectEnvImport)

	{ // Access-token group
		accessTokenGroup := projectEnvGroup.Group("/access-tokens")
		accessTokenGroup.GET("/:itemID", projectEnvSettingsHandler.GetAccessToken)
		accessTokenGroup.GET("", projectEnvSettingsHandler.ListAccessToken)
		accessTokenGroup.POST("", projectEnvSettingsHandler.CreateAccessToken)
		accessTokenGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateAccessToken)
		accessTokenGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateAccessTokenStatus)
		accessTokenGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteAccessToken)
		accessTokenGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeAccessToken))
	}

	{ // ACME DNS Provider group
		acmeDnsProviderGroup := projectEnvGroup.Group("/acme-dns-providers")
		acmeDnsProviderGroup.GET("/:itemID", projectEnvSettingsHandler.GetAcmeDnsProvider)
		acmeDnsProviderGroup.GET("", projectEnvSettingsHandler.ListAcmeDnsProvider)
		acmeDnsProviderGroup.POST("", projectEnvSettingsHandler.CreateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateAcmeDnsProvider)
		acmeDnsProviderGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateAcmeDnsProviderStatus)
		acmeDnsProviderGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteAcmeDnsProvider)
		acmeDnsProviderGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeAcmeDnsProvider))
	}

	{ // Audit log group
		auditLogGroup := projectEnvGroup.Group("/audit-logs")
		auditLogGroup.GET("", projectEnvHandler.ListAuditLog)
		auditLogGroup.GET("/types", projectEnvHandler.ListAuditLogTypes)
		auditLogGroup.GET("/:itemID", projectEnvHandler.GetAuditLog)
	}

	{ // Backup repository group
		backupSnapshotGroup := projectEnvGroup.Group("/backup-snapshots")
		backupSnapshotGroup.GET("", projectEnvSettingsHandler.ListBackupSnapshot)
		backupSnapshotGroup.GET("/:itemID", projectEnvSettingsHandler.GetBackupSnapshot)
		backupSnapshotGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteBackupSnapshot)
		backupSnapshotGroup.GET("/:itemID/entries", projectEnvSettingsHandler.ListBackupSnapshotEntries)
		backupSnapshotGroup.GET("/:itemID/download", projectEnvSettingsHandler.DownloadBackupSnapshotFile)
		backupSnapshotGroup.POST("/:itemID/restore", projectEnvSettingsHandler.RestoreBackupSnapshot)

		backupRepoGroup := projectEnvGroup.Group("/backup-repos")
		backupRepoGroup.GET("/:itemID", projectEnvSettingsHandler.GetBackupRepo)
		backupRepoGroup.GET("", projectEnvSettingsHandler.ListBackupRepo)
		backupRepoGroup.POST("", projectEnvSettingsHandler.CreateBackupRepo)
		backupRepoGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateBackupRepo)
		backupRepoGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateBackupRepoStatus)
		backupRepoGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteBackupRepo)
		backupRepoGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeBackupRepo))
		// Change repository password
		backupRepoGroup.PUT("/:itemID/password", projectEnvSettingsHandler.ChangeBackupRepoPassword)
		// Apply retention and reconcile stored snapshots
		backupRepoGroup.POST("/:itemID/cleanup", projectEnvSettingsHandler.CleanupBackupRepo)
		// Adopt options and snapshots changed on the repository outside the app
		backupRepoGroup.POST("/:itemID/sync", projectEnvSettingsHandler.SyncBackupRepo)
	}

	{ // Basic auth group
		basicAuthGroup := projectEnvGroup.Group("/basic-auth")
		basicAuthGroup.GET("/:itemID", projectEnvSettingsHandler.GetBasicAuth)
		basicAuthGroup.GET("", projectEnvSettingsHandler.ListBasicAuth)
		basicAuthGroup.POST("", projectEnvSettingsHandler.CreateBasicAuth)
		basicAuthGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateBasicAuth)
		basicAuthGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateBasicAuthStatus)
		basicAuthGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteBasicAuth)
		basicAuthGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeBasicAuth))
	}

	{ // Key auth group
		keyAuthGroup := projectEnvGroup.Group("/key-auth")
		keyAuthGroup.GET("/:itemID", projectEnvSettingsHandler.GetKeyAuth)
		keyAuthGroup.GET("", projectEnvSettingsHandler.ListKeyAuth)
		keyAuthGroup.POST("", projectEnvSettingsHandler.CreateKeyAuth)
		keyAuthGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateKeyAuth)
		keyAuthGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateKeyAuthStatus)
		keyAuthGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteKeyAuth)
		keyAuthGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeKeyAuth))
	}

	{ // Cloud storage group
		cloudStorageGroup := projectEnvGroup.Group("/cloud-storages")
		cloudStorageGroup.GET("/:itemID", projectEnvSettingsHandler.GetCloudStorage)
		cloudStorageGroup.GET("", projectEnvSettingsHandler.ListCloudStorage)
		cloudStorageGroup.POST("", projectEnvSettingsHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateCloudStorageStatus)
		cloudStorageGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteCloudStorage)
		cloudStorageGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeCloudStorage))
	}

	{ // Cluster network group
		networkGroup := projectEnvGroup.Group("/cluster-networks")
		networkGroup.GET("/:itemID", projectEnvSettingsHandler.GetClusterNetwork)
		networkGroup.GET("", projectEnvSettingsHandler.ListClusterNetwork)
		networkGroup.POST("", projectEnvSettingsHandler.CreateClusterNetwork)
		networkGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateClusterNetwork)
		networkGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateClusterNetworkStatus)
		networkGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteClusterNetwork)
		networkGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeClusterNetwork))
	}

	{ // Cluster volume group
		volumeGroup := projectEnvGroup.Group("/cluster-volumes")
		volumeGroup.GET("/:itemID", projectEnvSettingsHandler.GetClusterVolume)
		volumeGroup.GET("", projectEnvSettingsHandler.ListClusterVolume)
		volumeGroup.POST("", projectEnvSettingsHandler.CreateClusterVolume)
		volumeGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateClusterVolume)
		volumeGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateClusterVolumeStatus)
		volumeGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteClusterVolume)
		volumeGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeClusterVolume))
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
		commandPipeGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeCommandPipe))
	}

	{ // Scheduled jobs group: the env's own (job sequences) and, listed, its apps'
		schedJobGroup := projectEnvGroup.Group("/sched-jobs")
		schedJobGroup.GET("", projectEnvSettingsHandler.ListSchedJob)
		schedJobGroup.GET("/:itemID", projectEnvSettingsHandler.GetSchedJob)
		schedJobGroup.POST("", projectEnvSettingsHandler.CreateSchedJob)
		schedJobGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSchedJob)
		schedJobGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSchedJobStatus)
		schedJobGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSchedJob)
		schedJobGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeSchedJob))
		schedJobGroup.POST("/:itemID/exec", projectEnvSettingsHandler.ExecuteSchedJob)
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
		commandTemplateGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeCommandTemplate))
	}

	{ // Email group
		emailGroup := projectEnvGroup.Group("/emails")
		emailGroup.GET("/:itemID", projectEnvSettingsHandler.GetEmail)
		emailGroup.GET("", projectEnvSettingsHandler.ListEmail)
		emailGroup.POST("", projectEnvSettingsHandler.CreateEmail)
		emailGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateEmailStatus)
		emailGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteEmail)
		emailGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeEmail))
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
		imServiceGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeIMService))
	}

	{ // Notification group
		notificationGroup := projectEnvGroup.Group("/notifications")
		notificationGroup.GET("/:itemID", projectEnvSettingsHandler.GetNotification)
		notificationGroup.GET("", projectEnvSettingsHandler.ListNotification)
		notificationGroup.POST("", projectEnvSettingsHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateNotificationStatus)
		notificationGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteNotification)
		notificationGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeNotification))
	}

	{ // Registry auth group
		registryAuthGroup := projectEnvGroup.Group("/registry-auth")
		registryAuthGroup.GET("/:itemID", projectEnvSettingsHandler.GetRegistryAuth)
		registryAuthGroup.GET("", projectEnvSettingsHandler.ListRegistryAuth)
		registryAuthGroup.POST("", projectEnvSettingsHandler.CreateRegistryAuth)
		registryAuthGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateRegistryAuth)
		registryAuthGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateRegistryAuthStatus)
		registryAuthGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteRegistryAuth)
		registryAuthGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeRegistryAuth))
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
		secretGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeSecret))
	}

	{ // Config files
		configFileGroup := projectEnvGroup.Group("/config-files")
		configFileGroup.GET("", projectEnvSettingsHandler.ListConfigFile)
		configFileGroup.GET("/:itemID", projectEnvSettingsHandler.GetConfigFile)
		configFileGroup.POST("", projectEnvSettingsHandler.CreateConfigFile)
		configFileGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateConfigFile)
		configFileGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateConfigFileStatus)
		configFileGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteConfigFile)
		configFileGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeConfigFile))
	}

	{ // SSH key group
		sshKeyGroup := projectEnvGroup.Group("/ssh-keys")
		sshKeyGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSHKey)
		sshKeyGroup.GET("", projectEnvSettingsHandler.ListSSHKey)
		sshKeyGroup.POST("", projectEnvSettingsHandler.CreateSSHKey)
		sshKeyGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSHKey)
		sshKeyGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSHKeyStatus)
		sshKeyGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSHKey)
		sshKeyGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeSSHKey))
	}

	{ // SSL Cert group
		sslCertGroup := projectEnvGroup.Group("/ssl-certs")
		sslCertGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSLCert)
		sslCertGroup.GET("", projectEnvSettingsHandler.ListSSLCert)
		sslCertGroup.POST("", projectEnvSettingsHandler.CreateSSLCert)
		sslCertGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSLCert)
		sslCertGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSLCertStatus)
		sslCertGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSLCert)
		sslCertGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeSSLCert))
		sslCertGroup.POST("/:itemID/renew", projectEnvSettingsHandler.RenewSSLCert)
		sslCertGroup.GET("/:itemID/download", projectEnvSettingsHandler.DownloadSSLCertBundle)
	}

	{ // SSL Provider group
		sslProviderGroup := projectEnvGroup.Group("/ssl-providers")
		sslProviderGroup.GET("/:itemID", projectEnvSettingsHandler.GetSSLProvider)
		sslProviderGroup.GET("", projectEnvSettingsHandler.ListSSLProvider)
		sslProviderGroup.POST("", projectEnvSettingsHandler.CreateSSLProvider)
		sslProviderGroup.PUT("/:itemID", projectEnvSettingsHandler.UpdateSSLProvider)
		sslProviderGroup.PUT("/:itemID/status", projectEnvSettingsHandler.UpdateSSLProviderStatus)
		sslProviderGroup.DELETE("/:itemID", projectEnvSettingsHandler.DeleteSSLProvider)
		sslProviderGroup.GET("/:itemID/usages",
			projectEnvSettingsHandler.ProjectEnvSettingUsages(base.ResourceTypeSSLProvider))
	}

	{ // Task group
		taskGroup := projectEnvGroup.Group("/tasks")
		taskGroup.GET("", projectEnvHandler.ListTask)
		taskGroup.GET("/types", projectEnvHandler.ListTaskType)
		taskGroup.GET("/:itemID", projectEnvHandler.GetTask)
		taskGroup.GET("/:itemID/status", projectEnvHandler.GetTaskStatus)
		taskGroup.POST("/:itemID/cancel", projectEnvHandler.CancelTask)
		taskGroup.GET("/:itemID/logs", projectEnvHandler.GetTaskLogs)
		taskGroup.GET("/target-objects", projectEnvHandler.ListTaskTargetObject)
	}

	_ = s.registerAppRoutes(projectGroup, projectEnvGroup)
}
