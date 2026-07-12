package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type SetManagerNodesReq struct {
	Nodes basedto.ObjectIDSliceReq `json:"nodes"`
}

func NewSetManagerNodesReq() *SetManagerNodesReq {
	return &SetManagerNodesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetManagerNodesReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Nodes, true, 1, "nodes")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type SetManagerNodesResp struct {
	Meta *basedto.Meta `json:"meta"`
}
