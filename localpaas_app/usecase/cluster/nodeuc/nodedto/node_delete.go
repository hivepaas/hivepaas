package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteNodeReq struct {
	NodeID string `json:"-"`
	Force  bool   `json:"-" mapstructure:"force"`
}

func NewDeleteNodeReq() *DeleteNodeReq {
	return &DeleteNodeReq{}
}

func (req *DeleteNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.NodeID, true, 1, nodeIDMaxLen, "nodeID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
