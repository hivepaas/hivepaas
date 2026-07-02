package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetNodeInspectionReq struct {
	NodeID string `json:"-"`
}

func NewGetNodeInspectionReq() *GetNodeInspectionReq {
	return &GetNodeInspectionReq{}
}

func (req *GetNodeInspectionReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.NodeID, true, 1, nodeIDMaxLen, "nodeId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNodeInspectionResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data string        `json:"data"`
}
