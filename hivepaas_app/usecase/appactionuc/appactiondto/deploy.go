package appactiondto

import (
	"fmt"
	"regexp"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
)

const (
	imageNameMaxLen = 200
)

type DeployAppReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	ImageSource  *DeploymentImageSourceReq `json:"imageSource"`
	RepoSource   *DeploymentRepoSourceReq  `json:"repoSource"`
	ActiveMethod base.DeploymentMethod     `json:"activeMethod"`
	NoCache      bool                      `json:"noCache"`
	// ImageTags are extra tags for this deployment only, without the environment
	// prefix, which the build adds. Nothing stores them.
	ImageTags []string `json:"imageTags"`
	ChangeID  string   `json:"changeId"`
}

func (req *DeployAppReq) ApplyTo(setting *entity.AppDeploymentSettings) error {
	setting.ActiveMethod = gofn.Coalesce(req.ActiveMethod, setting.ActiveMethod)
	if setting.ActiveMethod == "" {
		return hperrors.Wrap(hperrors.ErrSettingMissing).WithParam("Name", "activeMethod")
	}
	switch setting.ActiveMethod {
	case base.DeploymentMethodImage:
		if err := req.ImageSource.ApplyTo(setting.ImageSource); err != nil {
			return hperrors.Wrap(err)
		}
	case base.DeploymentMethodRepo:
		if err := req.RepoSource.ApplyTo(setting.RepoSource); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

type DeploymentImageSourceReq struct {
	ImageTag string `json:"imageTag"`
}

func (req *DeploymentImageSourceReq) ApplyTo(setting *entity.DeploymentImageSource) error {
	if setting == nil || setting.Image == "" {
		return hperrors.Wrap(hperrors.ErrSettingMissing).WithParam("Name", "imageSource.image")
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
	RepoRef    string  `json:"repoRef"`
	CommitHash *string `json:"commitHash"`
}

func (req *DeploymentRepoSourceReq) ApplyTo(setting *entity.DeploymentRepoSource) error {
	if setting == nil || setting.RepoURL == "" {
		return hperrors.Wrap(hperrors.ErrSettingMissing).WithParam("Name", "repoSource.repoURL")
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
	}
	if setting.RepoRef == "" {
		return hperrors.Wrap(hperrors.ErrSettingMissing).WithParam("Name", "repoSource.repoRef")
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
func (req *DeployAppReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.ActiveMethod, false,
		base.AllDeploymentMethods, "activeMethod")...)
	validators = append(validators, req.ImageSource.validate("imageSource")...)
	validators = append(validators, req.RepoSource.validate("repoSource")...)
	validators = append(validators, req.validateImageTags()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// imageTagPattern is the OCI tag grammar, which has no place for a "/" or a ":",
// so a reference such as "api:v1" is refused here rather than pushed somewhere
// surprising. The length is checked separately, for a message that names it.
var imageTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)

// validateImageTags checks the tags this one deployment asked for. They carry no
// environment prefix: the build adds it, which is what keeps one environment's
// v1.4.0 from overwriting another's in the repository they share.
func (req *DeployAppReq) validateImageTags() (res []vld.Validator) {
	res = append(res, vld.SliceLen(req.ImageTags, 0, base.ImageMaxCustomTags).OnError(
		vld.SetField("imageTags", nil),
	))
	for i := range req.ImageTags {
		field := fmt.Sprintf("imageTags[%d]", i)
		res = append(res, vld.StrLen(&req.ImageTags[i], 1, base.ImageCustomTagMaxLen).OnError(
			vld.SetField(field, nil),
		))
		res = append(res, vld.StrMatch(&req.ImageTags[i], imageTagPattern).OnError(
			vld.SetField(field, nil),
		))
	}
	return res
}

type DeployAppResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *DeployAppDataResp `json:"data"`
}

type DeployAppDataResp struct {
	DeploymentID string `json:"deploymentId"`
}
