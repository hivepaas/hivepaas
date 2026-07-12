package appactiondto

import (
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
)

const (
	imageNameMaxLen = 200
)

type DeployAppReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	ImageSource  *DeploymentImageSourceReq `json:"imageSource"`
	RepoSource   *DeploymentRepoSourceReq  `json:"repoSource"`
	ActiveMethod base.DeploymentMethod     `json:"activeMethod"`
	NoCache      bool                      `json:"noCache"`
	ChangeID     string                    `json:"changeId"`
}

func (req *DeployAppReq) ApplyTo(setting *entity.AppDeploymentSettings) error {
	setting.ActiveMethod = gofn.Coalesce(req.ActiveMethod, setting.ActiveMethod)
	if setting.ActiveMethod == "" {
		return apperrors.Wrap(apperrors.ErrSettingMissing).WithNTParam("Name", "activeMethod")
	}
	switch setting.ActiveMethod {
	case base.DeploymentMethodImage:
		if err := req.ImageSource.ApplyTo(setting.ImageSource); err != nil {
			return apperrors.Wrap(err)
		}
	case base.DeploymentMethodRepo:
		if err := req.RepoSource.ApplyTo(setting.RepoSource); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

type DeploymentImageSourceReq struct {
	ImageTag string `json:"imageTag"`
}

func (req *DeploymentImageSourceReq) ApplyTo(setting *entity.DeploymentImageSource) error {
	if setting == nil || setting.Image == "" {
		return apperrors.Wrap(apperrors.ErrSettingMissing).WithNTParam("Name", "imageSource.image")
	}
	if req != nil {
		if req.ImageTag != "" {
			imageName, _, _ := strings.Cut(setting.Image, ":")
			setting.Image = imageName + ":" + req.ImageTag
		}
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
	res = append(res, basedto.ValidateStr(&req.ImageTag, false,
		1, imageNameMaxLen, field+"imageTag")...)
	return res
}

type DeploymentRepoSourceReq struct {
	RepoRef        string   `json:"repoRef"`
	CommitHash     *string  `json:"commitHash"`
	DockerfilePath string   `json:"dockerfilePath"` // for BuildToolDockerfile only
	ImageTags      []string `json:"imageTags"`
}

func (req *DeploymentRepoSourceReq) ApplyTo(setting *entity.DeploymentRepoSource) error {
	if setting == nil || setting.RepoURL == "" {
		return apperrors.Wrap(apperrors.ErrSettingMissing).WithNTParam("Name", "repoSource.repoURL")
	}
	if req != nil {
		setting.RepoRef = gofn.Coalesce(req.RepoRef, setting.RepoRef)
		// Normalize repo ref (currently supports git type only)
		switch setting.RepoType { //nolint:gocritic
		case base.RepoTypeGit:
			setting.RepoRef = string(githelper.NormalizeRepoRef(setting.RepoRef))
		}
		if req.CommitHash != nil {
			setting.CommitHash = *req.CommitHash
		}
		setting.DockerfilePath = gofn.Coalesce(req.DockerfilePath, setting.DockerfilePath)
		if req.ImageTags != nil {
			setting.ImageTags = req.ImageTags
		}
	}
	if setting.RepoRef == "" {
		return apperrors.Wrap(apperrors.ErrSettingMissing).WithNTParam("Name", "repoSource.repoRef")
	}
	return nil
}

func (req *DeploymentRepoSourceReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateGitCommitHash(req.CommitHash, false, field+"commitHash")...)
	return res
}

func NewDeployAppReq() *DeployAppReq {
	return &DeployAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeployAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
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
