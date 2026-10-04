package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentAppDeploymentSettingsVersion = 3
)

var _ = registerSettingParser(base.SettingTypeAppDeployment, &appDeploymentSettingsParser{})

type appDeploymentSettingsParser struct {
}

func (s *appDeploymentSettingsParser) New() SettingData {
	return &AppDeploymentSettings{}
}

type AppDeploymentSettings struct {
	ImageSource    *DeploymentImageSource    `json:"imageSource"`
	RepoSource     *DeploymentRepoSource     `json:"repoSource"`
	FunctionSource *DeploymentFunctionSource `json:"functionSource,omitempty"`
	ActiveMethod   base.DeploymentMethod     `json:"activeMethod"`

	// Entrypoint and Command are command lines, split by shell rules at each
	// deployment into the container's entrypoint and its arguments. Empty
	// leaves the image's.
	Entrypoint            string `json:"entrypoint,omitempty"`
	Command               string `json:"command,omitempty"`
	WorkingDir            string `json:"workingDir,omitempty"`
	PreDeploymentCommand  string `json:"preDeploymentCommand,omitempty"`
	PostDeploymentCommand string `json:"postDeploymentCommand,omitempty"`

	Notification *BaseEventNotification `json:"notification,omitempty"`
}

type DeploymentImageSource struct {
	Image        string   `json:"image"`
	RegistryAuth ObjectID `json:"registryAuth,omitzero"`
}

type DeploymentRepoSource struct {
	RepoType       base.RepoType         `json:"repoType"`
	RepoID         string                `json:"repoId"`
	RepoURL        string                `json:"repoURL"`
	RepoRef        string                `json:"repoRef"` // can be branch name, tag...
	CommitHash     string                `json:"commitHash,omitempty"`
	RepoOptions    DeploymentRepoOptions `json:"repoOptions"`
	Credentials    RepoCredentials       `json:"credentials,omitzero"` // id of github app/git token/ssh key setting
	Dockerfile     DeploymentDockerfile  `json:"dockerfile"`
	PushToRegistry ObjectID              `json:"pushToRegistry,omitzero"`
	// AutoDeploy is whether a push to RepoRef, received by a repo webhook,
	// deploys the app.
	AutoDeploy bool `json:"autoDeploy"`
}

type DeploymentRepoOptions struct {
	GitSubmodulesEnabled bool `json:"gitSubmodulesEnabled,omitempty"`
	GitLFSEnabled        bool `json:"gitLfsEnabled,omitempty"`
}

type DeploymentDockerfile struct {
	Source   base.DockerfileSource `json:"source"`
	Path     string                `json:"path"`
	Content  string                `json:"content,omitempty"`
	ScanPath string                `json:"scanPath,omitempty"`
}

type RepoCredentials struct {
	ID   string           `json:"id"`
	Type base.SettingType `json:"type"`
}

func (s *AppDeploymentSettings) GetType() base.SettingType {
	return base.SettingTypeAppDeployment
}

func (s *AppDeploymentSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.GetRegistryAuthIDs(), s.GetGitCredentialIDs()),
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *AppDeploymentSettings) GetRegistryAuthIDs() (res []string) {
	if s.ImageSource != nil && s.ImageSource.RegistryAuth.ID != "" {
		res = append(res, s.ImageSource.RegistryAuth.ID)
	}
	if s.RepoSource != nil && s.RepoSource.PushToRegistry.ID != "" {
		res = append(res, s.RepoSource.PushToRegistry.ID)
	}
	if s.FunctionSource != nil && s.FunctionSource.PushToRegistry.ID != "" {
		res = append(res, s.FunctionSource.PushToRegistry.ID)
	}
	res = gofn.ToSet(res)
	return
}

// ServiceRegistryAuthID is the registry credential the app's service pulls its
// image with, by the active method: the image's credential, or that of the
// registry a build pushes to. Empty when it pulls without one.
func (s *AppDeploymentSettings) ServiceRegistryAuthID() string {
	switch s.ActiveMethod {
	case base.DeploymentMethodImage:
		if s.ImageSource != nil {
			return s.ImageSource.RegistryAuth.ID
		}
	case base.DeploymentMethodRepo:
		if s.RepoSource != nil {
			return s.RepoSource.PushToRegistry.ID
		}
	case base.DeploymentMethodFunction:
		if s.FunctionSource != nil {
			return s.FunctionSource.PushToRegistry.ID
		}
	}
	return ""
}

func (s *AppDeploymentSettings) GetGitCredentialIDs() (res []string) {
	if s.RepoSource != nil && s.RepoSource.Credentials.ID != "" {
		res = append(res, s.RepoSource.Credentials.ID)
	}
	if s.FunctionSource != nil && s.FunctionSource.Code.Repo != nil && s.FunctionSource.Code.Repo.Credentials.ID != "" {
		res = append(res, s.FunctionSource.Code.Repo.Credentials.ID)
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppDeploymentSettings) GetResourceLinks(setting *Setting) []*ResLink {
	resLinks := s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)

	// Links repo ID (URL) to the current deployment
	if repoID, _, _ := s.sourceRepo(); repoID != "" && setting.ObjectID != "" {
		timeNow := timeutil.NowUTC()
		resLinks = append(resLinks, &ResLink{
			SrcType:   base.ResourceTypeSetting,
			SrcID:     setting.ID,
			DstType:   base.ResourceTypeRepo,
			DstID:     repoID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	return resLinks
}

// sourceRepo is the repository the app is built from, by the active method: its
// repo source, or its function's code in a repository.
func (s *AppDeploymentSettings) sourceRepo() (repoID, repoRef string, autoDeploy bool) {
	switch s.ActiveMethod {
	case base.DeploymentMethodRepo:
		if s.RepoSource != nil {
			return s.RepoSource.RepoID, s.RepoSource.RepoRef, s.RepoSource.AutoDeploy
		}
	case base.DeploymentMethodFunction:
		if s.FunctionSource != nil && s.FunctionSource.Code.Repo != nil {
			repo := s.FunctionSource.Code.Repo
			return repo.RepoID, repo.RepoRef, repo.AutoDeploy
		}
	case base.DeploymentMethodImage:
	}
	return "", "", false
}

// DeploysOnPush is whether a push to the repository and ref deploys the app: it
// is built from them, and set to deploy on push.
func (s *AppDeploymentSettings) DeploysOnPush(repoID, repoRef string) bool {
	sourceRepoID, sourceRepoRef, autoDeploy := s.sourceRepo()
	return autoDeploy && sourceRepoID != "" && sourceRepoID == repoID && sourceRepoRef == repoRef
}

// SetRepoCommitHash sets the commit the app's repository is built at, on the
// source of the active method. It tells whether that changed anything.
func (s *AppDeploymentSettings) SetRepoCommitHash(hash string) bool {
	var current *string
	switch s.ActiveMethod {
	case base.DeploymentMethodRepo:
		if s.RepoSource != nil {
			current = &s.RepoSource.CommitHash
		}
	case base.DeploymentMethodFunction:
		if s.FunctionSource != nil && s.FunctionSource.Code.Repo != nil {
			current = &s.FunctionSource.Code.Repo.CommitHash
		}
	case base.DeploymentMethodImage:
	}
	if current == nil || *current == hash {
		return false
	}
	*current = hash
	return true
}

func (s *Setting) AsAppDeploymentSettings() (*AppDeploymentSettings, error) {
	return parseSettingAs[*AppDeploymentSettings](s)
}

func (s *Setting) MustAsAppDeploymentSettings() *AppDeploymentSettings {
	return gofn.Must(s.AsAppDeploymentSettings())
}
