package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

type UpdateNodeReq struct {
	NodeID       string                  `json:"-"`
	Name         string                  `json:"name"`
	Labels       map[string]string       `json:"labels"`
	Role         docker.NodeRole         `json:"role"`
	Availability docker.NodeAvailability `json:"availability"`
	UpdateVer    int                     `json:"updateVer"`
}

func NewUpdateNodeReq() *UpdateNodeReq {
	return &UpdateNodeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.NodeID, true, 1, nodeIDMaxLen, "nodeId")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1, nodeNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Role, false, docker.AllNodeRoles, "role")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Availability, false, docker.AllNodeAvailabilities,
		"availability")...)

	// Validate node labels
	unallowedLabels := docker.ValidateUserLabels(req.Labels, true)
	if len(unallowedLabels) > 0 {
		validators = append(validators, vld.Must(false).OnError(
			vld.SetField("labels", nil),
			vld.SetCustomKey("ERR_VLD_DOCKER_LABEL_UNALLOWED"),
			vld.SetParam("Label", unallowedLabels[0]),
		))
	}

	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
