package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type CloneAppReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	SourceName   string         `json:"sourceName"`
	TargetName   string         `json:"targetName"`
	SourceEnv    string         `json:"sourceEnv"`
	TargetEnv    string         `json:"targetEnv"`
	SourceStatus base.AppStatus `json:"sourceStatus"`
	TargetStatus base.AppStatus `json:"targetStatus"`

	CloneConfigFiles        CloneConfigFilesReq        `json:"copyConfigFiles"`
	CloneDeploymentSettings CloneDeploymentSettingsReq `json:"copyDeploymentSettings"`
	CloneEnvVars            CloneEnvVarsReq            `json:"copyEnvVars"`
	CloneHealthChecks       CloneHealthChecksReq       `json:"copyHealthChecks"`
	CloneHttpSettings       CloneHttpSettingsReq       `json:"copyHttpSettings"`
	CloneSchedJobs          CloneSchedJobsReq          `json:"copySchedJobs"`
	CloneSecrets            CloneSecretsReq            `json:"copySecrets"`

	UpdateVer int `json:"updateVer"`
}

type CloneConfigFilesReq struct {
	Clone bool `json:"copy"`
}

type CloneDeploymentSettingsReq struct {
	Clone bool `json:"copy"`
}

type CloneEnvVarsReq struct {
	Clone bool `json:"copy"`
}

type CloneHealthChecksReq struct {
	Clone bool `json:"copy"`
}

type CloneHttpSettingsReq struct {
	Clone               bool                          `json:"copy"`
	CloneDomainSettings []*CloneHttpDomainSettingsReq `json:"copyDomainSettings"`
}

type CloneHttpDomainSettingsReq struct {
	SourceDomain  string              `json:"sourceDomain"`
	TargetDomain  string              `json:"targetDomain"`
	SourceSSLCert basedto.ObjectIDReq `json:"sourceSslCert"`
	TargetSSLCert basedto.ObjectIDReq `json:"targetSslCert"`
}

type CloneSchedJobsReq struct {
	Clone bool `json:"copy"`
}

type CloneSecretsReq struct {
	Clone bool `json:"copy"`
}

func NewCloneAppReq() *CloneAppReq {
	return &CloneAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CloneAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CloneAppResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
