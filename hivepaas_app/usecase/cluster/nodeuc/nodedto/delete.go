package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteNodeReq struct {
	settings.DeleteSettingReq
	Force bool `json:"-" mapstructure:"force"`
}

func NewDeleteNodeReq() *DeleteNodeReq {
	return &DeleteNodeReq{}
}

func (req *DeleteNodeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
