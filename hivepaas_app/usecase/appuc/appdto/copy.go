package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type CopyAppReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	SourceName   string         `json:"sourceName"`
	TargetName   string         `json:"targetName"`
	SourceEnv    string         `json:"sourceEnv"`
	TargetEnv    string         `json:"targetEnv"`
	SourceStatus base.AppStatus `json:"sourceStatus"`
	TargetStatus base.AppStatus `json:"targetStatus"`

	CopyConfigFiles        CopyConfigFilesReq        `json:"copyConfigFiles"`
	CopyDeploymentSettings CopyDeploymentSettingsReq `json:"copyDeploymentSettings"`
	CopyEnvVars            CopyEnvVarsReq            `json:"copyEnvVars"`
	CopyHealthChecks       CopyHealthChecksReq       `json:"copyHealthChecks"`
	CopyHttpSettings       CopyHttpSettingsReq       `json:"copyHttpSettings"`
	CopySchedJobs          CopySchedJobsReq          `json:"copySchedJobs"`
	CopySecrets            CopySecretsReq            `json:"copySecrets"`

	UpdateVer int `json:"updateVer"`
}

type CopyConfigFilesReq struct {
	Copy bool `json:"copy"`
}

type CopyDeploymentSettingsReq struct {
	Copy bool `json:"copy"`
}

type CopyEnvVarsReq struct {
	Copy bool `json:"copy"`
}

type CopyHealthChecksReq struct {
	Copy bool `json:"copy"`
}

type CopyHttpSettingsReq struct {
	Copy               bool                         `json:"copy"`
	CopyDomainSettings []*CopyHttpDomainSettingsReq `json:"copyDomainSettings"`
}

type CopyHttpDomainSettingsReq struct {
	SourceDomain  string              `json:"sourceDomain"`
	TargetDomain  string              `json:"targetDomain"`
	SourceSSLCert basedto.ObjectIDReq `json:"sourceSslCert"`
	TargetSSLCert basedto.ObjectIDReq `json:"targetSslCert"`
}

type CopySchedJobsReq struct {
	Copy bool `json:"copy"`
}

type CopySecretsReq struct {
	Copy bool `json:"copy"`
}

func NewCopyAppReq() *CopyAppReq {
	return &CopyAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CopyAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CopyAppResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
