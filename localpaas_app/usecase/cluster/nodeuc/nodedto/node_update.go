package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateNodeReq struct {
	NodeID       string                `json:"-"`
	Name         string                `json:"name"`
	Labels       map[string]string     `json:"labels"`
	Role         base.NodeRole         `json:"role"`
	Availability base.NodeAvailability `json:"availability"`
	UpdateVer    int                   `json:"updateVer"`
}

func NewUpdateNodeReq() *UpdateNodeReq {
	return &UpdateNodeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.NodeID, true, 1, nodeIDMaxLen, "nodeID")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1, nodeNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Role, false, base.AllNodeRoles, "role")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Availability, false, base.AllNodeAvailabilities,
		"availability")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
