package appdto

import (
	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/services/docker"
)

type GetAppDeploymentSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppDeploymentSettingsReq() *GetAppDeploymentSettingsReq {
	return &GetAppDeploymentSettingsReq{}
}

func (req *GetAppDeploymentSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppDeploymentSettingsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DeploymentSettingsResp `json:"data"`
}

type DeploymentSettingsResp struct {
	ImageSource  *DeploymentImageSourceResp `json:"imageSource,omitempty"`
	RepoSource   *DeploymentRepoSourceResp  `json:"repoSource,omitempty"`
	ActiveMethod base.DeploymentMethod      `json:"activeMethod"`

	Command               string `json:"command,omitempty"`
	WorkingDir            string `json:"workingDir,omitempty"`
	PreDeploymentCommand  string `json:"preDeploymentCommand,omitempty"`
	PostDeploymentCommand string `json:"postDeploymentCommand,omitempty"`

	Notification *basedto.BaseEventNotificationResp `json:"notification,omitempty"`

	UpdateVer int `json:"updateVer"`
}

type DeploymentImageSourceResp struct {
	Image        string                    `json:"image"`
	RegistryAuth *settings.BaseSettingResp `json:"registryAuth"`
}

type DeploymentRepoSourceResp struct {
	BuildTool      base.BuildTool            `json:"buildTool"`
	RepoType       base.RepoType             `json:"repoType"`
	RepoURL        string                    `json:"repoURL"`
	RepoRef        string                    `json:"repoRef"` // can be branch name, tag...
	Credentials    *settings.BaseSettingResp `json:"credentials"`
	DockerfilePath string                    `json:"dockerfilePath"` // for BuildToolDockerfile only
	ImageName      string                    `json:"imageName"`
	ImageTags      []string                  `json:"imageTags"`
	PushToRegistry *settings.BaseSettingResp `json:"pushToRegistry"`
}

type AppDeploymentSettingsTransformInput struct {
	App                *entity.App
	DeploymentSettings *entity.Setting
	ServiceSpec        *swarm.ServiceSpec
	RefObjects         *entity.RefObjects
}

func TransformDeploymentSettings(input *AppDeploymentSettingsTransformInput) (resp *DeploymentSettingsResp, err error) {
	resp = &DeploymentSettingsResp{}
	refObjects := input.RefObjects

	if input.ServiceSpec != nil && input.ServiceSpec.TaskTemplate.ContainerSpec != nil {
		resp.WorkingDir = input.ServiceSpec.TaskTemplate.ContainerSpec.Dir
		resp.Command = docker.ContainerCommandBuild(input.ServiceSpec.TaskTemplate.ContainerSpec.Command,
			input.ServiceSpec.TaskTemplate.ContainerSpec.Args)
	}

	var appDeploymentSettings *entity.AppDeploymentSettings
	if input.DeploymentSettings != nil {
		if err = copier.Copy(&resp, input.DeploymentSettings); err != nil {
			return nil, apperrors.Wrap(err)
		}
		appDeploymentSettings = input.DeploymentSettings.MustAsAppDeploymentSettings()
		if err = copier.Copy(&resp, appDeploymentSettings); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	if resp.ImageSource != nil { //nolint:nestif
		if resp.ImageSource.RegistryAuth != nil && resp.ImageSource.RegistryAuth.ID != "" {
			itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.ImageSource.RegistryAuth.ID])
			resp.ImageSource.RegistryAuth = itemResp
		}
	}
	if resp.RepoSource != nil { //nolint:nestif
		if resp.RepoSource.Credentials != nil && resp.RepoSource.Credentials.ID != "" {
			itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.RepoSource.Credentials.ID])
			resp.RepoSource.Credentials = itemResp
		}
		if resp.RepoSource.PushToRegistry != nil && resp.RepoSource.PushToRegistry.ID != "" {
			itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.RepoSource.PushToRegistry.ID])
			resp.RepoSource.PushToRegistry = itemResp
		}
	}

	if appDeploymentSettings != nil {
		resp.Notification = basedto.TransformBaseEventNotification(appDeploymentSettings.Notification, refObjects)
	}

	return resp, nil
}
