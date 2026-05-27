package base

type SubjectType string

const (
	SubjectTypeUser       SubjectType = "user"
	SubjectTypeCluster    SubjectType = "cluster"
	SubjectTypeNode       SubjectType = "node"
	SubjectTypeNetwork    SubjectType = "network"
	SubjectTypeDeployment SubjectType = "deployment"
	SubjectTypeProject    SubjectType = "project"
	SubjectTypeApp        SubjectType = "app"
)

type ResourceType string

const (
	ResourceTypeModule             ResourceType = "module"
	ResourceTypeUser               ResourceType = "user"
	ResourceTypeCluster            ResourceType = "cluster"
	ResourceTypeNode               ResourceType = "node"
	ResourceTypeNetwork            ResourceType = "network"
	ResourceTypeVolume             ResourceType = "volume"
	ResourceTypeImage              ResourceType = "image"
	ResourceTypeProject            ResourceType = "project"
	ResourceTypeApp                ResourceType = "app"
	ResourceTypeCloudStorage       ResourceType = "cloud-storage"
	ResourceTypeOAuth              ResourceType = "oauth"
	ResourceTypeSSHKey             ResourceType = "ssh-key"
	ResourceTypeAPIKey             ResourceType = "api-key"
	ResourceTypeSecret             ResourceType = "secret"
	ResourceTypeConfigFile         ResourceType = "config-file"
	ResourceTypeIMService          ResourceType = "im-service"
	ResourceTypeRegistryAuth       ResourceType = "registry-auth"
	ResourceTypeBasicAuth          ResourceType = "basic-auth"
	ResourceTypeSSLCert            ResourceType = "ssl-cert"
	ResourceTypeGithubApp          ResourceType = "github-app"
	ResourceTypeAccessToken        ResourceType = "access-token"
	ResourceTypeSchedJob           ResourceType = "sched-job"
	ResourceTypeHealthcheck        ResourceType = "healthcheck"
	ResourceTypeEmail              ResourceType = "email"
	ResourceTypeRepoWebhook        ResourceType = "repo-webhook"
	ResourceTypeSysError           ResourceType = "sys-error"
	ResourceTypeTask               ResourceType = "task"
	ResourceTypeNotification       ResourceType = "notification"
	ResourceTypeSystemCleanup      ResourceType = "system-cleanup"
	ResourceTypeSystemBackup       ResourceType = "system-backup"
	ResourceTypeSSLRenewal         ResourceType = "ssl-renewal"
	ResourceTypeFile               ResourceType = "file"
	ResourceTypeDomainSettings     ResourceType = "domain-settings"
	ResourceTypeStorageSettings    ResourceType = "storage-settings"
	ResourceTypeImageBuildSettings ResourceType = "image-build-settings"
	ResourceTypeLocalPaaSService   ResourceType = "localpaas-service"
	ResourceTypeTraefikService     ResourceType = "traefik-service"
)

type ResourceModule string

const (
	ResourceModuleSettings ResourceModule = "mod::settings"
	ResourceModuleCluster  ResourceModule = "mod::cluster"
	ResourceModuleUser     ResourceModule = "mod::user"
	ResourceModuleProject  ResourceModule = "mod::project"
	ResourceModuleSystem   ResourceModule = "mod::system"
)

var (
	AllResourceModules = []ResourceModule{ResourceModuleSettings, ResourceModuleUser,
		ResourceModuleCluster, ResourceModuleProject, ResourceModuleSystem}
)

type ActionType string

const (
	ActionTypeRead   ActionType = "read"
	ActionTypeWrite  ActionType = "write"
	ActionTypeDelete ActionType = "delete"
)

var (
	AllActionTypes = []ActionType{ActionTypeRead, ActionTypeWrite, ActionTypeDelete}
)

type AccessActions struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

func (a *AccessActions) Equal(other AccessActions) bool {
	if a.Delete && other.Delete {
		return true
	}
	if a.Delete == other.Delete && a.Write && other.Write {
		return true
	}
	return a.Read == other.Read && a.Write == other.Write && a.Delete == other.Delete
}

func (a *AccessActions) IsFullAccess() bool {
	return a.Read && a.Write && a.Delete
}

func (a *AccessActions) IsNoAccess() bool {
	return !a.Read && !a.Write && !a.Delete
}

func ActionAllowed(action ActionType, access AccessActions) bool {
	switch action {
	case ActionTypeRead:
		return access.Read || access.Write || access.Delete
	case ActionTypeWrite:
		return access.Write
	case ActionTypeDelete:
		return access.Delete
	}
	return false
}

type PermissionResource struct {
	SubjectType  SubjectType
	SubjectID    string
	ResourceType ResourceType
	ResourceID   string
}
