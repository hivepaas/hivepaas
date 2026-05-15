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
	SettingTypeCronJob            SettingType = "cron-job"
	SettingTypeHealthcheck        SettingType = "healthcheck"
	SettingTypeEmail              SettingType = "email"
	SettingTypeRepoWebhook        SettingType = "repo-webhook"
	SettingTypeNotification       SettingType = "notification"
	SettingTypeSystemCleanup      SettingType = "system-cleanup"
	SettingTypeSystemBackup       SettingType = "system-backup"
	SettingTypeSSLRenewal         SettingType = "ssl-renewal"
	SettingTypeFile               SettingType = "file"
	SettingTypeDomainSettings     SettingType = "domain-settings"
	SettingTypeStorageSettings    SettingType = "storage-settings"
	SettingTypeImageBuildSettings SettingType = "image-build-settings"
	SettingTypeLocalPaaSService   SettingType = "localpaas-service"
	SettingTypeTraefikService     SettingType = "traefik-service"
)

var (
	AllAppSettingTypes = []SettingType{SettingTypeApp, SettingTypeAppDeployment,
		SettingTypeAppHttp, SettingTypeEnvVar, SettingTypeSecret, SettingTypeCronJob, SettingTypeHealthcheck}

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

type SettingScopeType string

const (
	SettingScopeGlobal  SettingScopeType = ""
	SettingScopeUser    SettingScopeType = "user"
	SettingScopeProject SettingScopeType = "project"
	SettingScopeApp     SettingScopeType = "app"
)

type SettingScope struct {
	AppID            string
	ParentAppID      string
	ProjectID        string
	UserID           string
	NotRequireActive bool
}

func (s *SettingScope) ScopeType() SettingScopeType {
	switch {
	case s.AppID != "":
		return SettingScopeApp
	case s.ProjectID != "":
		return SettingScopeProject
	case s.UserID != "":
		return SettingScopeUser
	default:
		return SettingScopeGlobal
	}
}

func (s *SettingScope) IsGlobalScope() bool {
	return s.ScopeType() == SettingScopeGlobal
}

func (s *SettingScope) IsAppScope() bool {
	return s.AppID != ""
}

func (s *SettingScope) IsProjectScope() bool {
	return s.ProjectID != ""
}

func (s *SettingScope) IsUserScope() bool {
	return s.UserID != ""
}

func (s *SettingScope) MainObjectID() string {
	switch {
	case s.AppID != "":
		return s.AppID
	case s.ProjectID != "":
		return s.ProjectID
	case s.UserID != "":
		return s.UserID
	default:
		return ""
	}
}

func NewSettingScopeGlobal() *SettingScope {
	return &SettingScope{}
}

func NewSettingScopeApp(appID, projectID string) *SettingScope {
	return &SettingScope{
		AppID:     appID,
		ProjectID: projectID,
	}
}

func NewSettingScopeProject(projectID string) *SettingScope {
	return &SettingScope{
		ProjectID: projectID,
	}
}

func NewSettingScopeUser(userID string) *SettingScope {
	return &SettingScope{
		UserID: userID,
	}
}
