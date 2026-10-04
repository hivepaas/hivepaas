package appsettingsdto

import (
	"slices"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// healthcheckCommandMaxLen bounds a healthcheck's command.
const healthcheckCommandMaxLen = 4096

// healthcheckModes are the modes a healthcheck is written in.
var healthcheckModes = []docker.HealthcheckMode{docker.HealthcheckModeInherit, docker.HealthcheckModeNone,
	docker.HealthcheckModeCmd, docker.HealthcheckModeCmdShell}

type UpdateAppContainerSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	*BaseContainerSettings
	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppContainerSettingsReq() *UpdateAppContainerSettingsReq {
	return &UpdateAppContainerSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppContainerSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	if req.BaseContainerSettings == nil {
		validators = append(validators, basedto.ValidateCond(false, "image")...)
		return hperrors.NewValidationErrors(vld.Validate(validators...))
	}
	validators = append(validators, basedto.ValidateCommandLine(&req.Entrypoint, "entrypoint")...)
	validators = append(validators, basedto.ValidateCommandLine(&req.Command, "command")...)
	if check := req.Healthcheck; check != nil {
		validators = append(validators, basedto.ValidateCond(slices.Contains(healthcheckModes, check.Mode),
			"healthcheck.mode")...)
		if check.Enabled && (check.Mode == docker.HealthcheckModeCmd || check.Mode == docker.HealthcheckModeCmdShell) {
			validators = append(validators, basedto.ValidateStr(&check.Command, true, 1, healthcheckCommandMaxLen,
				"healthcheck.command")...)
		}
		if check.Mode == docker.HealthcheckModeCmd {
			validators = append(validators, basedto.ValidateCommandLine(&check.Command, "healthcheck.command")...)
		}
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppContainerSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
