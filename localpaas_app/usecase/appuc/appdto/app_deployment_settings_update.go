package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

const (
	imageNameMaxLen = 200
	repoRefMaxLen   = 200
)

type UpdateAppDeploymentSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	*DeploymentSettingsReq
}

type DeploymentSettingsReq struct {
	ImageSource  *DeploymentImageSourceReq `json:"imageSource"`
	RepoSource   *DeploymentRepoSourceReq  `json:"repoSource"`
	ActiveMethod base.DeploymentMethod     `json:"activeMethod"`

	Command               string `json:"command"`
	WorkingDir            string `json:"workingDir"`
	PreDeploymentCommand  string `json:"preDeploymentCommand"`
	PostDeploymentCommand string `json:"postDeploymentCommand"`

	Notification *basedto.BaseEventNotificationReq `json:"notification"`

	UpdateVer int `json:"updateVer"`
}

func (req *DeploymentSettingsReq) ToEntity() *entity.AppDeploymentSettings {
	return &entity.AppDeploymentSettings{
		ImageSource:  req.ImageSource.ToEntity(),
		RepoSource:   req.RepoSource.ToEntity(),
		ActiveMethod: req.ActiveMethod,

		Command:               req.Command,
		WorkingDir:            req.WorkingDir,
		PreDeploymentCommand:  req.PreDeploymentCommand,
		PostDeploymentCommand: req.PostDeploymentCommand,

		Notification: req.Notification.ToEntity(),
	}
}

type DeploymentImageSourceReq struct {
	Enabled      bool                `json:"enabled"`
	Image        string              `json:"image"`
	RegistryAuth basedto.ObjectIDReq `json:"registryAuth"`
}

func (req *DeploymentImageSourceReq) ToEntity() *entity.DeploymentImageSource {
	if req == nil {
		return nil
	}
	return &entity.DeploymentImageSource{
		Image:        req.Image,
		RegistryAuth: entity.ObjectID{ID: req.RegistryAuth.ID},
	}
}

func (req *DeploymentImageSourceReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Image, true, 1, imageNameMaxLen, field+"image")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.RegistryAuth, false, field+"registryAuth")...)
	return res
}

type DeploymentRepoSourceReq struct {
	Enabled        bool                `json:"enabled"`
	BuildTool      base.BuildTool      `json:"buildTool"`
	RepoType       base.RepoType       `json:"repoType"`
	RepoURL        string              `json:"repoURL"`
	RepoRef        string              `json:"repoRef"` // can be branch name, tag...
	Credentials    basedto.ObjectIDReq `json:"credentials"`
	DockerfilePath string              `json:"dockerfilePath"` // for BuildToolDockerfile only
	ImageName      string              `json:"imageName"`
	ImageTags      []string            `json:"imageTags"`
	PushToRegistry basedto.ObjectIDReq `json:"pushToRegistry"`
}

func (req *DeploymentRepoSourceReq) ToEntity() *entity.DeploymentRepoSource {
	if req == nil {
		return nil
	}
	return &entity.DeploymentRepoSource{
		BuildTool: req.BuildTool,
		RepoType:  req.RepoType,
		RepoURL:   req.RepoURL,
		RepoRef:   req.RepoRef,
		Credentials: entity.RepoCredentials{
			ID: req.Credentials.ID,
		},
		DockerfilePath: req.DockerfilePath,
		ImageName:      req.ImageName,
		ImageTags:      req.ImageTags,
		PushToRegistry: entity.ObjectID{ID: req.PushToRegistry.ID},
	}
}

func (req *DeploymentRepoSourceReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.BuildTool, true, base.AllBuildTools, field+"buildTool")...)
	res = append(res, basedto.ValidateStrIn(&req.RepoType, true, base.AllRepoTypes, field+"repoType")...)
	res = append(res, basedto.ValidateRepoURL(&req.RepoURL, true, field+"repoURL")...)
	res = append(res, basedto.ValidateStr(&req.RepoRef, false, 1, repoRefMaxLen, field+"repoRef")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Credentials, false, field+"credentials")...)
	res = append(res, basedto.ValidateStr(&req.ImageName, false, 1, base.ImageNameMaxLen, field+"imageName")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.PushToRegistry, false, field+"pushToRegistry")...)
	return res
}

func NewUpdateAppDeploymentSettingsReq() *UpdateAppDeploymentSettingsReq {
	return &UpdateAppDeploymentSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppDeploymentSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, req.ImageSource.validate("imageSource")...)
	validators = append(validators, req.RepoSource.validate("repoSource")...)
	validators = append(validators, basedto.ValidateStrIn(&req.ActiveMethod, true,
		base.AllDeploymentMethods, "activeMethod")...)
	validators = append(validators, req.Notification.Validate("notification")...)
	// TODO: add validation for deployment settings input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppDeploymentSettingsResp struct {
	Meta *basedto.Meta                        `json:"meta"`
	Data *UpdateAppDeploymentSettingsDataResp `json:"data"`
}

type UpdateAppDeploymentSettingsDataResp struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
