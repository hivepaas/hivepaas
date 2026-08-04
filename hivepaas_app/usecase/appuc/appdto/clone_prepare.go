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

type PrepareAppCloneReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewPrepareAppCloneReq() *PrepareAppCloneReq {
	return &PrepareAppCloneReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *PrepareAppCloneReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type PrepareAppCloneResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *PrepareAppCloneDataResp `json:"data"`
}

type PrepareAppCloneDataResp struct {
	SourceName   string         `json:"sourceName"`
	TargetName   string         `json:"targetName"`
	SourceEnv    string         `json:"sourceEnv"`
	TargetEnv    string         `json:"targetEnv"`
	SourceStatus base.AppStatus `json:"sourceStatus"`
	TargetStatus base.AppStatus `json:"targetStatus"`

	CloneConfigFiles        CloneConfigFilesResp        `json:"copyConfigFiles"`
	CloneDeploymentSettings CloneDeploymentSettingsResp `json:"copyDeploymentSettings"`
	CloneEnvVars            CloneEnvVarsResp            `json:"copyEnvVars"`
	CloneHealthChecks       CloneHealthChecksResp       `json:"copyHealthChecks"`
	CloneHttpSettings       CloneHttpSettingsResp       `json:"copyHttpSettings"`
	CloneSchedJobs          CloneSchedJobsResp          `json:"copySchedJobs"`
	CloneSecrets            CloneSecretsResp            `json:"copySecrets"`

	UpdateVer int `json:"updateVer"`
}

type CloneConfigFilesResp struct {
	Clone bool `json:"copy"`
}

type CloneDeploymentSettingsResp struct {
	Clone bool `json:"copy"`
}

type CloneEnvVarsResp struct {
	Clone bool `json:"copy"`
}

type CloneHealthChecksResp struct {
	Clone bool `json:"copy"`
}

type CloneHttpSettingsResp struct {
	Clone               bool                           `json:"copy"`
	CloneDomainSettings []*CloneHttpDomainSettingsResp `json:"copyDomainSettings"`
}

type CloneHttpDomainSettingsResp struct {
	SourceDomain  string                    `json:"sourceDomain"`
	TargetDomain  string                    `json:"targetDomain"`
	SourceSSLCert *settings.BaseSettingResp `json:"sourceSslCert"`
	TargetSSLCert *settings.BaseSettingResp `json:"targetSslCert"`
}

type CloneSchedJobsResp struct {
	Clone bool `json:"copy"`
}

type CloneSecretsResp struct {
	Clone bool `json:"copy"`
}

func TransformAppClonePreparationData(
	app *entity.App,
	refObjects *entity.RefObjects,
) (resp *PrepareAppCloneDataResp, err error) {
	resp = &PrepareAppCloneDataResp{
		SourceName:   app.Name,
		TargetName:   app.Name + " (copied)",
		SourceEnv:    app.ProjectEnv.Name,
		TargetEnv:    app.ProjectEnv.Name,
		SourceStatus: app.Status,
		TargetStatus: base.AppStatusActive,

		CloneConfigFiles:        CloneConfigFilesResp{Clone: true},
		CloneDeploymentSettings: CloneDeploymentSettingsResp{Clone: true},
		CloneEnvVars:            CloneEnvVarsResp{Clone: true},
		CloneHealthChecks:       CloneHealthChecksResp{Clone: true},
		CloneHttpSettings:       CloneHttpSettingsResp{Clone: true},
		CloneSchedJobs:          CloneSchedJobsResp{Clone: true},
		CloneSecrets:            CloneSecretsResp{Clone: true},

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
		resp.CloneHttpSettings.CloneDomainSettings = append(resp.CloneHttpSettings.CloneDomainSettings,
			&CloneHttpDomainSettingsResp{
				SourceDomain:  domain.Domain,
				TargetDomain:  "copied_" + domain.Domain,
				SourceSSLCert: sourceSslResp,
				TargetSSLCert: targetSslResp,
			})
	}

	return resp, nil
}
