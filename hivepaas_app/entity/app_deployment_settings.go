package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentAppDeploymentSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppDeployment, &appDeploymentSettingsParser{})

type appDeploymentSettingsParser struct {
}

func (s *appDeploymentSettingsParser) New() SettingData {
	return &AppDeploymentSettings{}
}

type AppDeploymentSettings struct {
	ImageSource  *DeploymentImageSource `json:"imageSource"`
	RepoSource   *DeploymentRepoSource  `json:"repoSource"`
	ActiveMethod base.DeploymentMethod  `json:"activeMethod"`
	NoCache      bool                   `json:"noCache,omitempty"`

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
	BuildTool      base.BuildTool        `json:"buildTool"`
	RepoType       base.RepoType         `json:"repoType"`
	RepoID         string                `json:"repoId"`
	RepoURL        string                `json:"repoURL"`
	RepoRef        string                `json:"repoRef"` // can be branch name, tag...
	CommitHash     string                `json:"commitHash,omitempty"`
	RepoOptions    DeploymentRepoOptions `json:"repoOptions"`
	Credentials    RepoCredentials       `json:"credentials,omitzero"`     // id of github app/git token/ssh key setting
	DockerfilePath string                `json:"dockerfilePath,omitempty"` // for BuildToolDockerfile only
	ImageName      string                `json:"imageName,omitempty"`
	ImageTags      []string              `json:"imageTags,omitempty"`
	PushToRegistry ObjectID              `json:"pushToRegistry,omitzero"`
}

type DeploymentRepoOptions struct {
	GitSubmodulesEnabled bool `json:"gitSubmodulesEnabled,omitempty"`
	GitLFSEnabled        bool `json:"gitLfsEnabled,omitempty"`
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
	res = gofn.ToSet(res)
	return
}

func (s *AppDeploymentSettings) GetGitCredentialIDs() (res []string) {
	if s.RepoSource != nil && s.RepoSource.Credentials.ID != "" {
		res = append(res, s.RepoSource.Credentials.ID)
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppDeploymentSettings) CalcResLinks(setting *Setting) []*ResLink {
	resLinks := s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)

	// Links repo ID (URL) to the current deployment
	if s.ActiveMethod == base.DeploymentMethodRepo && s.RepoSource != nil && setting.ObjectID != "" {
		timeNow := timeutil.NowUTC()
		resLinks = append(resLinks, &ResLink{
			SrcType:   base.ResourceTypeSetting,
			SrcID:     setting.ID,
			DstType:   base.ResourceTypeRepo,
			DstID:     s.RepoSource.RepoID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	return resLinks
}

func (s *AppDeploymentSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppDeploymentSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppDeploymentSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppDeploymentSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAppDeploymentSettings() (*AppDeploymentSettings, error) {
	return parseSettingAs[*AppDeploymentSettings](s)
}

func (s *Setting) MustAsAppDeploymentSettings() *AppDeploymentSettings {
	return gofn.Must(s.AsAppDeploymentSettings())
}
