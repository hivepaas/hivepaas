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
// Access that is turned off keeps what it allowed, for turning it on again, and
// host mode keeps the proxy's policy for going back to it.
type AppDockerAPISettingsResp struct {
	Enabled bool `json:"enabled"`
	// Mode is proxy or host, never empty.
	Mode       string              `json:"mode"`
	Images     []string            `json:"images"`
	SharedDirs []string            `json:"sharedDirs"`
	Networks   []string            `json:"networks"`
	Allow      []string            `json:"allow"`
	Limits     *AppDockerAPILimits `json:"limits"`
	// DefaultLimits are what a limit of zero stands for.
	DefaultLimits *AppDockerAPILimits `json:"defaultLimits"`
	// HostMode says whether the caller may choose host mode.
	HostMode  *AppDockerAPIHostModeResp `json:"hostMode"`
	UpdateVer int                       `json:"updateVer"`
}

// AppDockerAPIHostModeResp says whether the caller may give the app the node's
// own socket, and when not, what is missing: "switch", the privileged-apps
// switch, or "admin", an administrator. The switch is named first: only an
// administrator can turn it on.
type AppDockerAPIHostModeResp struct {
	Available bool   `json:"available"`
	BlockedBy string `json:"blockedBy"`
}

// AppDockerAPILimits bound the app's children: how many there may be at once,
// and the most one may use. Zero is the default.
type AppDockerAPILimits struct {
	Containers int           `json:"containers"`
	Memory     unit.DataSize `json:"memory" swaggertype:"string"`
	CPUs       float64       `json:"cpus"`
}

// TransformAppDockerAPISettings is what the screen shows of an app's setting,
// nil when it has none, to a caller host mode is blocked for as blockedBy says.
func TransformAppDockerAPISettings(setting *entity.Setting, blockedBy string) (*AppDockerAPISettingsResp, error) {
	resp := &AppDockerAPISettingsResp{
		Mode:   entity.DockerAPIModeProxy,
		Images: []string{}, SharedDirs: []string{}, Networks: []string{}, Allow: []string{},
		Limits: &AppDockerAPILimits{},
		DefaultLimits: &AppDockerAPILimits{
			Containers: dockerapiservice.DefaultContainers,
			Memory:     dockerapiservice.DefaultMemory,
			CPUs:       dockerapiservice.DefaultCPUs,
		},
		HostMode: &AppDockerAPIHostModeResp{Available: blockedBy == "", BlockedBy: blockedBy},
	}
	if setting == nil {
		return resp, nil
	}
	access, err := setting.AsAppDockerAPISettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Enabled = setting.Status == base.SettingStatusActive
	if access.IsHostMode() {
		resp.Mode = entity.DockerAPIModeHost
	}
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
