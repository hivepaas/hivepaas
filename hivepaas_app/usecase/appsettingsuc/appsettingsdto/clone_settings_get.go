package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetAppCloneSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppCloneSettingsReq() *GetAppCloneSettingsReq {
	return &GetAppCloneSettingsReq{}
}

func (req *GetAppCloneSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppCloneSettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *AppCloneSettingsResp `json:"data"`
}

type AppCloneSettingsResp struct {
	TargetName     string         `json:"targetName,omitempty"`
	TargetEnv      string         `json:"targetEnv,omitempty"`
	TargetStatus   base.AppStatus `json:"targetStatus,omitempty"`
	TargetReplicas int            `json:"targetReplicas,omitempty"`

	CloneDeploymentSettings bool                                 `json:"cloneDeploymentSettings,omitempty"`
	CloneRoutingSettings    bool                                 `json:"cloneRoutingSettings,omitempty"`
	CloneRoutingDomains     []*AppCloneRoutingDomainSettingsResp `json:"cloneRoutingDomains,omitempty" copy:"-"`

	CloneVolumes    bool     `json:"cloneVolumes"`
	CloneVolumeData bool     `json:"cloneVolumeData"`
	LiveVolumeClone bool     `json:"liveVolumeClone"`
	IncludedVolumes []string `json:"includedVolumes,omitempty"`
	ExcludedVolumes []string `json:"excludedVolumes,omitempty"`

	CloneEnvVars      bool `json:"cloneEnvVars,omitempty"`
	CloneSecrets      bool `json:"cloneSecrets,omitempty"`
	CloneConfigFiles  bool `json:"cloneConfigFiles,omitempty"`
	ClonePeriodicJobs bool `json:"clonePeriodicJobs,omitempty"`
	CloneSchedJobs    bool `json:"cloneSchedJobs,omitempty"`

	CommandPipes []*settings.BaseSettingResp        `json:"commandPipes,omitempty"`
	Notification *basedto.BaseEventNotificationResp `json:"notification,omitempty"`

	UpdateVer int `json:"updateVer"`
}

type AppCloneRoutingDomainSettingsResp struct {
	SourceDomain  string                    `json:"sourceDomain"`
	TargetDomain  string                    `json:"targetDomain"`
	SourceSSLCert *settings.BaseSettingResp `json:"sourceSslCert,omitempty"`
	TargetSSLCert *settings.BaseSettingResp `json:"targetSslCert,omitempty"`
}

type AppCloneSettingsTransformInput struct {
	App              *entity.App
	AppCloneSettings *entity.AppCloneSettings
	RefObjects       *entity.RefObjects
	UpdateVer        int
}

func TransformAppCloneSettings(input *AppCloneSettingsTransformInput) (*AppCloneSettingsResp, error) {
	resp := &AppCloneSettingsResp{
		UpdateVer: input.UpdateVer,
	}

	appCloneSettings := input.AppCloneSettings
	if input.RefObjects == nil {
		input.RefObjects = &entity.RefObjects{}
	}
	refObjects := input.RefObjects

	if appCloneSettings != nil {
		if err := copier.Copy(&resp, &appCloneSettings); err != nil {
			return nil, apperrors.Wrap(err)
		}

		resp.CloneRoutingDomains = nil
		for _, domain := range appCloneSettings.CloneRoutingDomains {
			domainResp := &AppCloneRoutingDomainSettingsResp{
				SourceDomain: domain.SourceDomain,
				TargetDomain: domain.TargetDomain,
			}
			if domain.SourceSSLCert.ID != "" {
				certResp, _ := settings.TransformSettingBase(refObjects.RefSettings[domain.SourceSSLCert.ID])
				if certResp == nil {
					certResp = settings.NewMissingSetting(domain.SourceSSLCert.ID, base.SettingTypeSSLCert)
				}
				domainResp.SourceSSLCert = certResp
			}
			if domain.TargetSSLCert.ID != "" {
				certResp, _ := settings.TransformSettingBase(refObjects.RefSettings[domain.TargetSSLCert.ID])
				if certResp == nil {
					certResp = settings.NewMissingSetting(domain.TargetSSLCert.ID, base.SettingTypeSSLCert)
				}
				domainResp.TargetSSLCert = certResp
			}
			resp.CloneRoutingDomains = append(resp.CloneRoutingDomains, domainResp)
		}

		for _, pipe := range appCloneSettings.CommandPipes {
			pipeResp, _ := settings.TransformSettingBase(refObjects.RefSettings[pipe.ID])
			if pipeResp == nil {
				pipeResp = settings.NewMissingSetting(pipe.ID, base.SettingTypeCommandPipe)
			}
			resp.CommandPipes = append(resp.CommandPipes, pipeResp)
		}

		resp.Notification = basedto.TransformBaseEventNotification(appCloneSettings.Notification, refObjects)
	}

	return resp, nil
}
