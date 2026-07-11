package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteNodeReq struct {
	settings.DeleteSettingReq
	Force bool `json:"-" mapstructure:"force"`
}

func NewDeleteNodeReq() *DeleteNodeReq {
	return &DeleteNodeReq{}
}

func (req *DeleteNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
