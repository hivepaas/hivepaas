package appdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type PrepareAppCopyReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewPrepareAppCopyReq() *PrepareAppCopyReq {
	return &PrepareAppCopyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *PrepareAppCopyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type PrepareAppCopyResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *PrepareAppCopyDataResp `json:"data"`
}

type PrepareAppCopyDataResp struct {
	SourceName   string         `json:"sourceName"`
	TargetName   string         `json:"targetName"`
	SourceEnv    string         `json:"sourceEnv"`
	TargetEnv    string         `json:"targetEnv"`
	SourceStatus base.AppStatus `json:"sourceStatus"`
	TargetStatus base.AppStatus `json:"targetStatus"`

	CopyConfigFiles        CopyConfigFilesResp        `json:"copyConfigFiles"`
	CopyDeploymentSettings CopyDeploymentSettingsResp `json:"copyDeploymentSettings"`
	CopyEnvVars            CopyEnvVarsResp            `json:"copyEnvVars"`
	CopyHealthChecks       CopyHealthChecksResp       `json:"copyHealthChecks"`
	CopyHttpSettings       CopyHttpSettingsResp       `json:"copyHttpSettings"`
	CopySchedJobs          CopySchedJobsResp          `json:"copySchedJobs"`
	CopySecrets            CopySecretsResp            `json:"copySecrets"`

	UpdateVer int `json:"updateVer"`
}

type CopyConfigFilesResp struct {
	Copy bool `json:"copy"`
}

type CopyDeploymentSettingsResp struct {
	Copy bool `json:"copy"`
}

type CopyEnvVarsResp struct {
	Copy bool `json:"copy"`
}

type CopyHealthChecksResp struct {
	Copy bool `json:"copy"`
}

type CopyHttpSettingsResp struct {
	Copy               bool                          `json:"copy"`
	CopyDomainSettings []*CopyHttpDomainSettingsResp `json:"copyDomainSettings"`
}

type CopyHttpDomainSettingsResp struct {
	SourceDomain  string                    `json:"sourceDomain"`
	TargetDomain  string                    `json:"targetDomain"`
	SourceSSLCert *settings.BaseSettingResp `json:"sourceSslCert"`
	TargetSSLCert *settings.BaseSettingResp `json:"targetSslCert"`
}

type CopySchedJobsResp struct {
	Copy bool `json:"copy"`
}

type CopySecretsResp struct {
	Copy bool `json:"copy"`
}

func TransformAppCopyPreparationData(
	app *entity.App,
	refObjects *entity.RefObjects,
) (resp *PrepareAppCopyDataResp, err error) {
	resp = &PrepareAppCopyDataResp{
		SourceName:   app.Name,
		TargetName:   app.Name + " (copied)",
		SourceEnv:    app.Env,
		TargetEnv:    app.Env,
		SourceStatus: app.Status,
		TargetStatus: base.AppStatusActive,

		CopyConfigFiles:        CopyConfigFilesResp{Copy: true},
		CopyDeploymentSettings: CopyDeploymentSettingsResp{Copy: true},
		CopyEnvVars:            CopyEnvVarsResp{Copy: true},
		CopyHealthChecks:       CopyHealthChecksResp{Copy: true},
		CopyHttpSettings:       CopyHttpSettingsResp{Copy: true},
		CopySchedJobs:          CopySchedJobsResp{Copy: true},
		CopySecrets:            CopySecretsResp{Copy: true},

		UpdateVer: app.UpdateVer,
	}

	httpSetting := app.GetSettingByType(base.SettingTypeAppHttp)
	httpSettings := httpSetting.MustAsAppHttpSettings()
	for _, domain := range httpSettings.Domains {
		sslCert := refObjects.RefSettings[domain.SSLCert.ID]
		sourceSslResp, _ := settings.TransformSettingBase(sslCert)
		targetSslResp := sourceSslResp
		if sslCert != nil && !strings.HasPrefix(sslCert.MustAsSSLCert().Domain, "*.") {
			targetSslResp = nil
		}
		resp.CopyHttpSettings.CopyDomainSettings = append(resp.CopyHttpSettings.CopyDomainSettings,
			&CopyHttpDomainSettingsResp{
				SourceDomain:  domain.Domain,
				TargetDomain:  "copied_" + domain.Domain,
				SourceSSLCert: sourceSslResp,
				TargetSSLCert: targetSslResp,
			})
	}

	return resp, nil
}
