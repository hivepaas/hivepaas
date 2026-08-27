package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type SetManagerNodesReq struct {
	Nodes basedto.ObjectIDSliceReq `json:"nodes"`
}

func NewSetManagerNodesReq() *SetManagerNodesReq {
	return &SetManagerNodesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetManagerNodesReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Nodes, true, 1, "nodes")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type SetManagerNodesResp struct {
	Meta *basedto.Meta `json:"meta"`
}
