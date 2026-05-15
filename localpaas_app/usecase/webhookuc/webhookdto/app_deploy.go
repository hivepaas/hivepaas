package webhookdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

const (
	imageNameMaxLen = 200
	appTokenMaxLen  = 100
)

type DeployAppReq struct {
	AppToken string `json:"-"`

	ImageSource         *DeploymentImageSourceReq `json:"imageSource"`
	RepoSource          *DeploymentRepoSourceReq  `json:"repoSource"`
	ActiveMethod        base.DeploymentMethod     `json:"activeMethod"`
	DeploymentTriggerID string                    `json:"deploymentTriggerId"`
}

func (req *DeployAppReq) ApplyTo(setting *entity.AppDeploymentSettings) error {
	setting.ActiveMethod = gofn.Coalesce(req.ActiveMethod, setting.ActiveMethod)
	if err := req.ImageSource.ApplyTo(setting.ImageSource); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.RepoSource.ApplyTo(setting.RepoSource); err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type DeploymentImageSourceReq struct {
	Image string `json:"image"`
}

func (req *DeploymentImageSourceReq) ApplyTo(setting *entity.DeploymentImageSource) error {
	if req != nil {
		setting.Image = gofn.Coalesce(req.Image, setting.Image)
	}
	if setting.Image == "" {
		return apperrors.New(apperrors.ErrSettingMissing).WithNTParam("Name", "imageSource.image")
	}
	return nil
}

func (req *DeploymentImageSourceReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Image, true, 1, imageNameMaxLen, field+"image")...)
	return res
}

type DeploymentRepoSourceReq struct {
	RepoRef        string   `json:"repoRef"`        // can be branch name, tag...
	DockerfilePath string   `json:"dockerfilePath"` // for BuildToolDockerfile only
	ImageTags      []string `json:"imageTags"`
}

func (req *DeploymentRepoSourceReq) ApplyTo(setting *entity.DeploymentRepoSource) error {
	if req != nil {
		setting.RepoRef = gofn.Coalesce(req.RepoRef, setting.RepoRef)
		setting.DockerfilePath = gofn.Coalesce(req.DockerfilePath, setting.DockerfilePath)
		if len(req.ImageTags) > 0 {
			setting.ImageTags = req.ImageTags
		}
	}
	if setting.RepoRef == "" {
		return apperrors.New(apperrors.ErrSettingMissing).WithNTParam("Name", "repoSource.repoRef")
	}
	return nil
}

// nolint
func (req *DeploymentRepoSourceReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	// TODO: add implementation
	return res
}

func NewDeployAppReq() *DeployAppReq {
	return &DeployAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeployAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.AppToken, true,
		1, appTokenMaxLen, "appToken")...)
	validators = append(validators, basedto.ValidateStrIn(&req.ActiveMethod, false,
		base.AllDeploymentMethods, "activeMethod")...)
	validators = append(validators, req.ImageSource.validate("imageSource")...)
	validators = append(validators, req.RepoSource.validate("repoSource")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeployAppResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *DeployAppDataResp `json:"data"`
}

type DeployAppDataResp struct {
	DeploymentID string `json:"deploymentId"`
}
