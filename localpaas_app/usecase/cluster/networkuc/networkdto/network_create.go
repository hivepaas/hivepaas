package networkdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	nameMaxLen = 64
)

type CreateNetworkReq struct {
	ProjectID       string            `json:"projectID"`
	AvailInProjects bool              `json:"availableInProjects"`
	Name            string            `json:"name"`
	Driver          string            `json:"driver"`
	EnableIPv4      bool              `json:"enableIPv4"`
	EnableIPv6      bool              `json:"enableIPv6"`
	Internal        bool              `json:"internal"`
	Attachable      bool              `json:"attachable"`
	Ingress         bool              `json:"ingress"`
	Options         map[string]string `json:"options"`
	Labels          map[string]string `json:"labels"`
}

func NewCreateNetworkReq() *CreateNetworkReq {
	return &CreateNetworkReq{}
}

func (req *CreateNetworkReq) ModifyRequest() error {
	req.Driver = gofn.Coalesce(req.Driver, docker.NetworkDriverOverlay)
	if !req.EnableIPv4 && !req.EnableIPv6 {
		req.EnableIPv4 = true
	}
	return nil
}

func (req *CreateNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, nameMaxLen, "name")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateNetworkResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
