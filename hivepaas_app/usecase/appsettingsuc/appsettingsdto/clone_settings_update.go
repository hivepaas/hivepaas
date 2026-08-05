package appsettingsdto

import (
	"fmt"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	domainMaxLen = 100
)

type UpdateAppCloneSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	UpdateVer    int    `json:"updateVer"`

	TargetName     string         `json:"targetName"`
	TargetEnv      string         `json:"targetEnv"`
	TargetStatus   base.AppStatus `json:"targetStatus"`
	TargetReplicas int            `json:"targetReplicas"`

	CloneDeploymentSettings bool                             `json:"cloneDeploymentSettings"`
	CloneHttpSettings       bool                             `json:"cloneHttpSettings"`
	CloneHttpDomains        []*AppCloneHttpDomainSettingsReq `json:"cloneHttpDomains"`

	CloneVolumes    bool     `json:"cloneVolumes"`
	CloneVolumeData bool     `json:"cloneVolumeData"`
	LiveVolumeClone bool     `json:"liveVolumeClone"`
	IncludedVolumes []string `json:"includedVolumes"`
	ExcludedVolumes []string `json:"excludedVolumes"`

	CloneEnvVars      bool `json:"cloneEnvVars"`
	CloneSecrets      bool `json:"cloneSecrets"`
	CloneConfigFiles  bool `json:"cloneConfigFiles"`
	ClonePeriodicJobs bool `json:"clonePeriodicJobs"`
	CloneSchedJobs    bool `json:"cloneSchedJobs"`

	CommandPipes basedto.ObjectIDSliceReq          `json:"commandPipes"`
	Notification *basedto.BaseEventNotificationReq `json:"notification"`
}

func NewUpdateAppCloneSettingsReq() *UpdateAppCloneSettingsReq {
	return &UpdateAppCloneSettingsReq{}
}

func (req *UpdateAppCloneSettingsReq) ModifyRequest() error {
	req.TargetName = strings.TrimSpace(req.TargetName)
	req.TargetEnv = strings.TrimSpace(req.TargetEnv)
	return nil
}

func (req *UpdateAppCloneSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.TargetStatus, false,
		base.AllAppStatuses, "targetStatus")...)

	for i, domain := range req.CloneHttpDomains {
		field := fmt.Sprintf("cloneHttpDomains[%d].", i)
		validators = append(validators, basedto.ValidateDomain(&domain.SourceDomain, true, domainMaxLen,
			false, field+"sourceDomain")...)
		validators = append(validators, basedto.ValidateDomain(&domain.TargetDomain, true, domainMaxLen,
			false, field+"targetDomain")...)
	}

	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateAppCloneSettingsReq) ToEntity() *entity.AppCloneSettings {
	return &entity.AppCloneSettings{
		TargetName:     req.TargetName,
		TargetEnv:      req.TargetEnv,
		TargetStatus:   req.TargetStatus,
		TargetReplicas: req.TargetReplicas,

		CloneDeploymentSettings: req.CloneDeploymentSettings,
		CloneHttpSettings:       req.CloneHttpSettings,
		CloneHttpDomains: gofn.MapSlice(req.CloneHttpDomains,
			func(d *AppCloneHttpDomainSettingsReq) *entity.AppCloneHttpDomainSettings {
				return d.ToEntity()
			}),

		CloneVolumes:    req.CloneVolumes,
		CloneVolumeData: req.CloneVolumeData,
		LiveVolumeClone: req.LiveVolumeClone,
		IncludedVolumes: req.IncludedVolumes,
		ExcludedVolumes: req.ExcludedVolumes,

		CloneEnvVars:      req.CloneEnvVars,
		CloneSecrets:      req.CloneSecrets,
		CloneConfigFiles:  req.CloneConfigFiles,
		ClonePeriodicJobs: req.ClonePeriodicJobs,
		CloneSchedJobs:    req.CloneSchedJobs,

		CommandPipes: req.CommandPipes.ToEntity(),
		Notification: req.Notification.ToEntity(),
	}
}

type AppCloneHttpDomainSettingsReq struct {
	SourceDomain  string              `json:"sourceDomain"`
	TargetDomain  string              `json:"targetDomain"`
	SourceSSLCert basedto.ObjectIDReq `json:"sourceSslCert"`
	TargetSSLCert basedto.ObjectIDReq `json:"targetSslCert"`
}

func (req *AppCloneHttpDomainSettingsReq) ToEntity() *entity.AppCloneHttpDomainSettings {
	if req == nil {
		return nil
	}
	return &entity.AppCloneHttpDomainSettings{
		SourceDomain:  req.SourceDomain,
		TargetDomain:  req.TargetDomain,
		SourceSSLCert: *req.SourceSSLCert.ToEntity(),
		TargetSSLCert: *req.TargetSSLCert.ToEntity(),
	}
}

type UpdateAppCloneSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
