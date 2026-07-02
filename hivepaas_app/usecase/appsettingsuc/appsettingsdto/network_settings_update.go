package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppNetworkSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	NetworkAttachments []*NetworkAttachment `json:"networkAttachments"`
	HostsFileEntries   []*HostsFileEntry    `json:"hostsFileEntries"`
	DNSConfig          *DNSConfig           `json:"dnsConfig"`
	EndpointSpec       *EndpointSpec        `json:"endpointSpec"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppNetworkSettingsReq() *UpdateAppNetworkSettingsReq {
	return &UpdateAppNetworkSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppNetworkSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppNetworkSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppNetworkSettingsDataResp `json:"data"`
}

type UpdateAppNetworkSettingsDataResp struct {
}
