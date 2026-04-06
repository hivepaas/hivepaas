package networkdto

import (
	"time"

	"github.com/docker/docker/api/types/network"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	networkIDMaxLen = 100
)

type GetNetworkReq struct {
	NetworkID string `json:"-"`
	ProjectID string `json:"-"`
}

func NewGetNetworkReq() *GetNetworkReq {
	return &GetNetworkReq{}
}

func (req *GetNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: network id is docker id normally
	validators = append(validators, basedto.ValidateStr(&req.NetworkID, true, 1, networkIDMaxLen, "networkID")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, false, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *NetworkResp  `json:"data"`
}

type NetworkResp struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	AvailInProjects bool              `json:"availableInProjects"`
	Driver          string            `json:"driver"`
	Internal        bool              `json:"internal"`
	Attachable      bool              `json:"attachable"`
	Ingress         bool              `json:"ingress"`
	EnableIPv4      bool              `json:"enableIPv4"`
	EnableIPv6      bool              `json:"enableIPv6"`
	Options         map[string]string `json:"options"`
	Labels          map[string]string `json:"labels"`
	CreatedAt       time.Time         `json:"createdAt"`
}

func TransformNetworkInspection(net *network.Inspect) *NetworkResp {
	return &NetworkResp{
		ID:              net.ID,
		Name:            net.Name,
		AvailInProjects: net.Labels[docker.StackLabelNamespace] == "",
		Driver:          net.Driver,
		Internal:        net.Internal,
		Attachable:      net.Attachable,
		Ingress:         net.Ingress,
		EnableIPv4:      net.EnableIPv4,
		EnableIPv6:      net.EnableIPv6,
		Options:         net.Options,
		Labels:          net.Labels,
		CreatedAt:       net.Created,
	}
}

func TransformNetwork(net *network.Summary) *NetworkResp {
	return &NetworkResp{
		ID:              net.ID,
		Name:            net.Name,
		AvailInProjects: net.Labels[docker.StackLabelNamespace] == "",
		Driver:          net.Driver,
		Internal:        net.Internal,
		Attachable:      net.Attachable,
		Ingress:         net.Ingress,
		EnableIPv4:      net.EnableIPv4,
		EnableIPv6:      net.EnableIPv6,
		Options:         net.Options,
		Labels:          net.Labels,
		CreatedAt:       net.Created,
	}
}
