package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

type GetAppDockerAPISettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppDockerAPISettingsReq() *GetAppDockerAPISettingsReq {
	return &GetAppDockerAPISettingsReq{}
}

func (req *GetAppDockerAPISettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppDockerAPISettingsResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *AppDockerAPISettingsResp `json:"data"`
}

// AppDockerAPISettingsResp is an app's Docker API access as its screen shows it.
// Access that is turned off keeps what it allowed, for turning it on again.
type AppDockerAPISettingsResp struct {
	Enabled    bool                `json:"enabled"`
	Images     []string            `json:"images"`
	SharedDirs []string            `json:"sharedDirs"`
	Networks   []string            `json:"networks"`
	Allow      []string            `json:"allow"`
	Limits     *AppDockerAPILimits `json:"limits"`
	// DefaultLimits are what a limit of zero stands for.
	DefaultLimits *AppDockerAPILimits `json:"defaultLimits"`
	UpdateVer     int                 `json:"updateVer"`
}

// AppDockerAPILimits bound the app's children: how many there may be at once,
// and the most one may use. Zero is the default.
type AppDockerAPILimits struct {
	Containers int           `json:"containers"`
	Memory     unit.DataSize `json:"memory" swaggertype:"string"`
	CPUs       float64       `json:"cpus"`
}

// TransformAppDockerAPISettings is what the screen shows of an app's setting,
// nil when it has none.
func TransformAppDockerAPISettings(setting *entity.Setting) (*AppDockerAPISettingsResp, error) {
	resp := &AppDockerAPISettingsResp{
		Images: []string{}, SharedDirs: []string{}, Networks: []string{}, Allow: []string{},
		Limits: &AppDockerAPILimits{},
		DefaultLimits: &AppDockerAPILimits{
			Containers: dockerapiservice.DefaultContainers,
			Memory:     dockerapiservice.DefaultMemory,
			CPUs:       dockerapiservice.DefaultCPUs,
		},
	}
	if setting == nil {
		return resp, nil
	}
	access, err := setting.AsAppDockerAPISettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Enabled = setting.Status == base.SettingStatusActive
	resp.Images = append(resp.Images, access.Images...)
	resp.SharedDirs = append(resp.SharedDirs, access.SharedDirs...)
	resp.Networks = append(resp.Networks, access.Networks...)
	resp.Allow = append(resp.Allow, access.Allow...)
	resp.Limits = &AppDockerAPILimits{
		Containers: access.Limits.Containers, Memory: access.Limits.Memory, CPUs: access.Limits.CPUs,
	}
	resp.UpdateVer = setting.UpdateVer
	return resp, nil
}
