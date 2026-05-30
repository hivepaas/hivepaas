package base

import "github.com/tiendc/gofn"

const (
	SettingNameMaxLen = 100
)

type SettingType string

const (
	SettingTypeProject            SettingType = "project"
	SettingTypeProjectEnvs        SettingType = "project-envs"
	SettingTypeApp                SettingType = "app"
	SettingTypeAppDeployment      SettingType = "app-deployment"
	SettingTypeAppHttp            SettingType = "app-http"
	SettingTypeEnvVar             SettingType = "env-var"
	SettingTypeSecret             SettingType = "secret"
	SettingTypeConfigFile         SettingType = "config-file"
	SettingTypeCloudStorage       SettingType = "cloud-storage"
	SettingTypeOAuth              SettingType = "oauth"
	SettingTypeSSHKey             SettingType = "ssh-key"
	SettingTypeAPIKey             SettingType = "api-key"
	SettingTypeIMService          SettingType = "im-service"
	SettingTypeRegistryAuth       SettingType = "registry-auth"
	SettingTypeBasicAuth          SettingType = "basic-auth"
	SettingTypeSSLCert            SettingType = "ssl-cert"
	SettingTypeGithubApp          SettingType = "github-app"
	SettingTypeAccessToken        SettingType = "access-token"
	SettingTypeSchedJob           SettingType = "sched-job"
	SettingTypeHealthcheck        SettingType = "healthcheck"
	SettingTypeEmail              SettingType = "email"
	SettingTypeRepoWebhook        SettingType = "repo-webhook"
	SettingTypeNotification       SettingType = "notification"
	SettingTypeSystemCleanup      SettingType = "system-cleanup"
	SettingTypeSystemBackup       SettingType = "system-backup"
	SettingTypeSSLRenewal         SettingType = "ssl-renewal"
	SettingTypeDomainSettings     SettingType = "domain-settings"
	SettingTypeStorageSettings    SettingType = "storage-settings"
	SettingTypeImageBuildSettings SettingType = "image-build-settings"
	SettingTypeLocalPaaSService   SettingType = "localpaas-service"
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
