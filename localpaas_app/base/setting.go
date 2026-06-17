package base

import "github.com/tiendc/gofn"

const (
	SettingNameMaxLen = 100
)

type SettingType string

const (
	SettingTypeAccessToken        SettingType = "access-token"
	SettingTypeAcmeDnsProvider    SettingType = "acme-dns-provider"
	SettingTypeAPIKey             SettingType = "api-key"
	SettingTypeApp                SettingType = "app"
	SettingTypeAppDeployment      SettingType = "app-deployment"
	SettingTypeAppFeatures        SettingType = "app-features"
	SettingTypeAppHttp            SettingType = "app-http"
	SettingTypeBasicAuth          SettingType = "basic-auth"
	SettingTypeCloudStorage       SettingType = "cloud-storage"
	SettingTypeConfigFile         SettingType = "config-file"
	SettingTypeDomainSettings     SettingType = "domain-settings"
	SettingTypeEmail              SettingType = "email"
	SettingTypeEnvVar             SettingType = "env-var"
	SettingTypeGithubApp          SettingType = "github-app"
	SettingTypeHealthcheck        SettingType = "healthcheck"
	SettingTypeImageBuildSettings SettingType = "image-build-settings"
	SettingTypeIMService          SettingType = "im-service"
	SettingTypeLocalPaaSService   SettingType = "localpaas-service"
	SettingTypeNotification       SettingType = "notification"
	SettingTypeOAuth              SettingType = "oauth"
	SettingTypeProject            SettingType = "project"
	SettingTypeProjectEnvs        SettingType = "project-envs"
	SettingTypeRegistryAuth       SettingType = "registry-auth"
	SettingTypeRepoWebhook        SettingType = "repo-webhook"
	SettingTypeSSHKey             SettingType = "ssh-key"
	SettingTypeSSLCert            SettingType = "ssl-cert"
	SettingTypeSSLProvider        SettingType = "ssl-provider"
	SettingTypeSSLRenewal         SettingType = "ssl-renewal"
	SettingTypeSchedJob           SettingType = "sched-job"
	SettingTypeSecret             SettingType = "secret"
	SettingTypeStorageSettings    SettingType = "storage-settings"
	SettingTypeSystemBackup       SettingType = "system-backup"
	SettingTypeSystemCleanup      SettingType = "system-cleanup"
	SettingTypeTraefikService     SettingType = "traefik-service"
)

var (
	AllAppSettingTypes = []SettingType{SettingTypeApp, SettingTypeAppDeployment,
		SettingTypeAppHttp, SettingTypeEnvVar, SettingTypeSecret, SettingTypeSchedJob, SettingTypeHealthcheck}

	AllProjectSettingTypes = []SettingType{SettingTypeProject, SettingTypeEnvVar, SettingTypeSecret}
)

type SettingStatus string

const (
	SettingStatusActive   SettingStatus = "active"
	SettingStatusPending  SettingStatus = "pending"
	SettingStatusDisabled SettingStatus = "disabled"
	SettingStatusExpired  SettingStatus = "expired"
)

var (
	AllSettingStatuses = []SettingStatus{SettingStatusActive, SettingStatusPending, SettingStatusDisabled,
		SettingStatusExpired}
	AllSettingSettableStatuses = gofn.Drop(AllSettingStatuses, SettingStatusExpired)
)
