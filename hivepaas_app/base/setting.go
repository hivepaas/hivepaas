package base

import "github.com/tiendc/gofn"

type SettingType string

const (
	SettingTypeAccessToken     SettingType = "access-token"
	SettingTypeAcmeDnsProvider SettingType = "acme-dns-provider"
	SettingTypeAPIKey          SettingType = "api-key"
	SettingTypeApp             SettingType = "app"
	SettingTypeAppClone        SettingType = "app-clone"
	SettingTypeAppDeployment   SettingType = "app-deployment"
	SettingTypeAppFeatures     SettingType = "app-features"
	SettingTypeAppPlacement    SettingType = "app-placement"
	SettingTypeAppRouting      SettingType = "app-routing"
	SettingTypeBackupRepo      SettingType = "backup-repo"
	SettingTypeBackupSnapshot  SettingType = "backup-snapshot"
	SettingTypeBasicAuth       SettingType = "basic-auth"
	SettingTypeCloudStorage    SettingType = "cloud-storage"
	SettingTypeClusterNetwork  SettingType = "cluster-network"
	SettingTypeClusterNode     SettingType = "cluster-node"
	SettingTypeClusterVolume   SettingType = "cluster-volume"
	SettingTypeCommandPipe     SettingType = "command-pipe"
	SettingTypeCommandTemplate SettingType = "command-template"
	SettingTypeConfigFile      SettingType = "config-file"
	SettingTypeDomainSettings  SettingType = "domain-settings"
	SettingTypeEmail           SettingType = "email"
	SettingTypeEnvVar          SettingType = "env-var"
	SettingTypeGithubApp       SettingType = "github-app"
	SettingTypeImageBuild      SettingType = "image-build"
	SettingTypeIMService       SettingType = "im-service"
	SettingTypeHivePaaSService SettingType = "hivepaas-service"
	SettingTypeNotification    SettingType = "notification"
	SettingTypeOAuth           SettingType = "oauth"
	SettingTypePeriodicJob     SettingType = "periodic-job"
	SettingTypeProject         SettingType = "project"
	SettingTypeRegistryAuth    SettingType = "registry-auth"
	SettingTypeRepoWebhook     SettingType = "repo-webhook"
	SettingTypeScript          SettingType = "script"
	SettingTypeSSHKey          SettingType = "ssh-key"
	SettingTypeSSLCert         SettingType = "ssl-cert"
	SettingTypeSSLProvider     SettingType = "ssl-provider"
	SettingTypeSSLRenewal      SettingType = "ssl-renewal"
	SettingTypeSchedJob        SettingType = "sched-job"
	SettingTypeSecret          SettingType = "secret"
	SettingTypeStorageSettings SettingType = "storage-settings"
	SettingTypeSystemBackup    SettingType = "system-backup"
	SettingTypeSystemCleanup   SettingType = "system-cleanup"
	SettingTypeTraefikService  SettingType = "traefik-service"
)

var (
	AllAppSettingTypes = []SettingType{SettingTypeApp, SettingTypeAppDeployment,
		SettingTypeAppRouting, SettingTypeEnvVar, SettingTypeSecret, SettingTypeConfigFile,
		SettingTypeSchedJob, SettingTypePeriodicJob}

	AllProjectSettingTypes = []SettingType{SettingTypeProject, SettingTypeEnvVar, SettingTypeSecret}
)

type SettingStatus string

const (
	SettingStatusActive   SettingStatus = "active"
	SettingStatusPending  SettingStatus = "pending"
	SettingStatusDisabled SettingStatus = "disabled"
	SettingStatusExpired  SettingStatus = "expired"
	SettingStatusMissing  SettingStatus = "missing" // this is not used in DB
)

var (
	AllSettingStatuses = []SettingStatus{SettingStatusActive, SettingStatusPending, SettingStatusDisabled,
		SettingStatusExpired}
	AllSettingSettableStatuses = gofn.Drop(AllSettingStatuses, SettingStatusExpired)
)
