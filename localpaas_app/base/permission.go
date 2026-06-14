package base

type SubjectType string

const (
	SubjectTypeApp        SubjectType = "app"
	SubjectTypeCluster    SubjectType = "cluster"
	SubjectTypeDeployment SubjectType = "deployment"
	SubjectTypeNetwork    SubjectType = "network"
	SubjectTypeNode       SubjectType = "node"
	SubjectTypeProject    SubjectType = "project"
	SubjectTypeUser       SubjectType = "user"
)

type ResourceType string

const (
	ResourceTypeAccessToken        ResourceType = "access-token"
	ResourceTypeAPIKey             ResourceType = "api-key"
	ResourceTypeApp                ResourceType = "app"
	ResourceTypeBasicAuth          ResourceType = "basic-auth"
	ResourceTypeCloudStorage       ResourceType = "cloud-storage"
	ResourceTypeCluster            ResourceType = "cluster"
	ResourceTypeConfigFile         ResourceType = "config-file"
	ResourceTypeDomain             ResourceType = "domain"
	ResourceTypeDomainSettings     ResourceType = "domain-settings"
	ResourceTypeEmail              ResourceType = "email"
	ResourceTypeGithubApp          ResourceType = "github-app"
	ResourceTypeHealthcheck        ResourceType = "healthcheck"
	ResourceTypeImage              ResourceType = "image"
	ResourceTypeImageBuildSettings ResourceType = "image-build-settings"
	ResourceTypeIMService          ResourceType = "im-service"
	ResourceTypeLocalPaaSService   ResourceType = "localpaas-service"
	ResourceTypeModule             ResourceType = "module"
	ResourceTypeNetwork            ResourceType = "network"
	ResourceTypeNode               ResourceType = "node"
	ResourceTypeNotification       ResourceType = "notification"
	ResourceTypeOAuth              ResourceType = "oauth"
	ResourceTypeProject            ResourceType = "project"
	ResourceTypeRegistryAuth       ResourceType = "registry-auth"
	ResourceTypeRepo               ResourceType = "repo"
	ResourceTypeRepoWebhook        ResourceType = "repo-webhook"
	ResourceTypeSchedJob           ResourceType = "sched-job"
	ResourceTypeSecret             ResourceType = "secret"
	ResourceTypeSSHKey             ResourceType = "ssh-key"
	ResourceTypeSSLCert            ResourceType = "ssl-cert"
	ResourceTypeSSLProvider        ResourceType = "ssl-provider"
	ResourceTypeSSLRenewal         ResourceType = "ssl-renewal"
	ResourceTypeStorageSettings    ResourceType = "storage-settings"
	ResourceTypeSysError           ResourceType = "sys-error"
	ResourceTypeSystemBackup       ResourceType = "system-backup"
	ResourceTypeSystemCleanup      ResourceType = "system-cleanup"
	ResourceTypeTask               ResourceType = "task"
	ResourceTypeTraefikService     ResourceType = "traefik-service"
	ResourceTypeUser               ResourceType = "user"
	ResourceTypeVolume             ResourceType = "volume"
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
	ActionTypeRead    ActionType = "read"
	ActionTypeExecute ActionType = "execute"
	ActionTypeWrite   ActionType = "write"
	ActionTypeDelete  ActionType = "delete"
)

var (
	AllActionTypes = []ActionType{ActionTypeRead, ActionTypeExecute, ActionTypeWrite, ActionTypeDelete}
)

type AccessActions struct {
	Read  bool `json:"read"`
	Exec  bool `json:"execute"`
	Write bool `json:"write"`
	Del   bool `json:"delete"`
}

func (a *AccessActions) Equal(other AccessActions) bool {
	return a.Read == other.Read && a.Exec == other.Exec && a.Write == other.Write && a.Del == other.Del
}

func (a *AccessActions) IsFullAccess() bool {
	return a.Read && a.Exec && a.Write && a.Del
}

func (a *AccessActions) IsNoAccess() bool {
	return !a.Read && !a.Exec && !a.Write && !a.Del
}

func (a *AccessActions) GetAllowedActions() (allowed []ActionType) {
	if a.Read {
		allowed = append(allowed, ActionTypeRead)
	}
	if a.Exec {
		allowed = append(allowed, ActionTypeExecute)
	}
	if a.Write {
		allowed = append(allowed, ActionTypeWrite)
	}
	if a.Del {
		allowed = append(allowed, ActionTypeDelete)
	}
	return allowed
}

func (a *AccessActions) Allows(action ActionType) bool {
	if a == nil {
		return false
	}
	switch action {
	case ActionTypeRead:
		return a.Read
	case ActionTypeExecute:
		return a.Exec
	case ActionTypeWrite:
		return a.Write
	case ActionTypeDelete:
		return a.Del
	}
	return false
}

func (a *AccessActions) AllowsAll(actions []ActionType) bool {
	if a == nil {
		return false
	}
	for _, action := range actions {
		if !a.Allows(action) {
			return false
		}
	}
	return true
}

func (a *AccessActions) AllowsAny(actions []ActionType) bool {
	if a == nil {
		return false
	}
	if len(actions) == 0 {
		return true
	}
	for _, action := range actions {
		if !a.Allows(action) {
			return true
		}
	}
	return false
}

func (a *AccessActions) Reset(r, x, w, d bool) {
	a.Read = r
	a.Exec = x
	a.Write = w
	a.Del = d
}

func NewAccessActions(r, x, w, d bool) AccessActions {
	return AccessActions{
		Read:  r,
		Exec:  x,
		Write: w,
		Del:   d,
	}
}

func NewFullAccessActions() AccessActions {
	return AccessActions{
		Read:  true,
		Exec:  true,
		Write: true,
		Del:   true,
	}
}

type PermissionResource struct {
	SubjectType  SubjectType
	SubjectID    string
	ResourceType ResourceType
	ResourceID   string
}
