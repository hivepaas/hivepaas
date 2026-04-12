package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

type UpdateAppContainerSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	*ContainerSpec

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppContainerSettingsReq() *UpdateAppContainerSettingsReq {
	return &UpdateAppContainerSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppContainerSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)

	// Validate service labels
	unallowedLabels := docker.ValidateUserLabels(req.Labels, true)
	if len(unallowedLabels) > 0 {
		validators = append(validators, vld.Must(false).OnError(
			vld.SetField("labels", nil),
			vld.SetCustomKey("ERR_VLD_DOCKER_LABEL_UNALLOWED"),
			vld.SetParam("Label", unallowedLabels[0]),
		))
	}

	// TODO: validate other input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppContainerSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
