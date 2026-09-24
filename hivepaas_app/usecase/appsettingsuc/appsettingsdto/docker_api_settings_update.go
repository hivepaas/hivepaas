package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateAppDockerAPISettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	// Enabled off takes the access away and keeps what it allowed, for turning
	// it on again: the other fields are not read.
	Enabled    bool                `json:"enabled"`
	Images     []string            `json:"images"`
	SharedDirs []string            `json:"sharedDirs"`
	Networks   []string            `json:"networks"`
	Allow      []string            `json:"allow"`
	Limits     *AppDockerAPILimits `json:"limits"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppDockerAPISettingsReq() *UpdateAppDockerAPISettingsReq {
	return &UpdateAppDockerAPISettingsReq{}
}

// Validate implements interface basedto.ReqValidator. What the access may say
// is checked by the use case, against the app's service.
func (req *UpdateAppDockerAPISettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ToEntity is the access the request asks for.
func (req *UpdateAppDockerAPISettingsReq) ToEntity() *entity.AppDockerAPISettings {
	access := &entity.AppDockerAPISettings{
		Images: req.Images, SharedDirs: req.SharedDirs, Networks: req.Networks, Allow: req.Allow,
	}
	if req.Limits != nil {
		access.Limits = entity.AppDockerAPILimits{
			Containers: req.Limits.Containers, Memory: req.Limits.Memory, CPUs: req.Limits.CPUs,
		}
	}
	return access
}

type UpdateAppDockerAPISettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
