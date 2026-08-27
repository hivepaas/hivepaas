package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateAppNetworkSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

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
func (req *UpdateAppNetworkSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppNetworkSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppNetworkSettingsDataResp `json:"data"`
}

type UpdateAppNetworkSettingsDataResp struct {
}
