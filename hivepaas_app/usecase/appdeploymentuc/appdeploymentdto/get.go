package appdeploymentdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
)

const (
	commitHashShortLen = 7
)

type GetDeploymentReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	DeploymentID string `json:"-"`
}

func NewGetDeploymentReq() *GetDeploymentReq {
	return &GetDeploymentReq{}
}

func (req *GetDeploymentReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *DeploymentResp `json:"data"`
}

type DeploymentResp struct {
	ID        string                        `json:"id"`
	Status    base.DeploymentStatus         `json:"status"`
	UpdateVer int                           `json:"updateVer"`
	Settings  *entity.AppDeploymentSettings `json:"settings"`
	Trigger   *DeploymentTriggerResp        `json:"trigger"`
	Output    *DeploymentOutputResp         `json:"output"`

	StartedAt *time.Time `json:"startedAt" copy:",nilonzero"`
	EndedAt   *time.Time `json:"endedAt" copy:",nilonzero"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type DeploymentTriggerResp struct {
	Source     base.DeploymentTriggerSource `json:"source"`
	SourceUser *basedto.UserBaseResp        `json:"sourceUser,omitempty"`
}

type DeploymentOutputResp struct {
	Error           string   `json:"error,omitempty"`
	CommitHash      string   `json:"commitHash,omitempty"`
	CommitHashShort string   `json:"commitHashShort,omitempty"`
	CommitURL       string   `json:"commitURL,omitempty"`
	CommitTitle     string   `json:"commitTitle,omitempty"`
	CommitMessage   string   `json:"commitMessage,omitempty"`
	CommitAuthor    string   `json:"commitAuthor,omitempty"`
	ImageTags       []string `json:"imageTags,omitempty"`
}

func TransformDeployment(
	deployment *entity.Deployment,
	input *DeploymentTransformInput,
) (resp *DeploymentResp, err error) {
	if err = copier.Copy(&resp, &deployment); err != nil {
		return nil, hperrors.Wrap(err)
	}

	deploymentInfo := input.DeploymentInfoMap[deployment.ID]
	if deploymentInfo != nil {
		resp.Status = deploymentInfo.Status
		if deploymentInfo.Status == base.DeploymentStatusInProgress {
			resp.StartedAt = &deploymentInfo.StartedAt
		}
	}

	if deployment.Trigger != nil && (deployment.Trigger.Source == base.DeploymentTriggerSourceUser ||
		deployment.Trigger.Source == base.DeploymentTriggerSourceAPI) {
		triggerUser := input.TriggerUserMap[deployment.Trigger.SourceID]
		if triggerUser != nil {
			resp.Trigger.SourceUser = basedto.TransformUserBase(triggerUser)
		} else {
			resp.Trigger.SourceUser = basedto.NewMissingUser(deployment.Trigger.SourceID)
		}
	}

	if resp.Output != nil {
		resp.Output.Error = strings.Join(deployment.Output.Errors, "\n")
		resp.Output.CommitHashShort = resp.Output.CommitHash
		if len(resp.Output.CommitHashShort) > commitHashShortLen { // shorten to some characters if possible
			resp.Output.CommitHashShort = resp.Output.CommitHashShort[:commitHashShortLen]
		}
		if resp.Output.CommitHash != "" && deployment.Settings.RepoSource != nil {
			repoURL := deployment.Settings.RepoSource.RepoID
			resp.Output.CommitURL = githelper.GetCommitHttpsUrl(repoURL, resp.Output.CommitHash)
		}
	}

	return resp, nil
}
