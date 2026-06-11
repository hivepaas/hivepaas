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
	ResourceTypeDomain             ResourceType = "domain"
	ResourceTypeRepo               ResourceType = "repo"
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
	ResourceTypeSSLProvider        ResourceType = "ssl-provider"
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
