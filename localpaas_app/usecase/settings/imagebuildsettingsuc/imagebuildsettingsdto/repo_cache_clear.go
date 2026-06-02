package imagebuildsettingsdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ClearRepoCacheReq struct {
	Scope *base.ObjectScope `json:"-" mapstructure:"-"`
}

func NewClearRepoCacheReq() *ClearRepoCacheReq {
	return &ClearRepoCacheReq{}
}

func (req *ClearRepoCacheReq) Validate() apperrors.ValidationErrors {
	// TODO: add validation
	return nil
}

type ClearRepoCacheResp struct {
	Meta *basedto.Meta `json:"meta"`
}
