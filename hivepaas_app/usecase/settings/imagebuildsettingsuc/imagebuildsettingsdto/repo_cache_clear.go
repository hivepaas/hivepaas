package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ClearRepoCacheReq struct {
	Scope *base.ObjectScope `json:"-" mapstructure:"-"`
}

func NewClearRepoCacheReq() *ClearRepoCacheReq {
	return &ClearRepoCacheReq{}
}

func (req *ClearRepoCacheReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateCond(req.Scope.IsValid(), "params")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ClearRepoCacheResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *ClearRepoCacheDataResp `json:"data"`
}

type ClearRepoCacheDataResp struct {
	FilesDeleted   int    `json:"filesDeleted"`
	SpaceReclaimed uint64 `json:"spaceReclaimed"`
}
