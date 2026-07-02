package imagebuildsettingsdto

import (
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
	// TODO: add validation
	return nil
}

type ClearRepoCacheResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *ClearRepoCacheDataResp `json:"data"`
}

type ClearRepoCacheDataResp struct {
	FilesDeleted   int    `json:"filesDeleted"`
	SpaceReclaimed uint64 `json:"spaceReclaimed"`
}
