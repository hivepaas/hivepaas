package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate env var input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppNetworkSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppNetworkSettingsDataResp `json:"data"`
}

type UpdateAppNetworkSettingsDataResp struct {
}
