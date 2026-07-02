package builddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ClearBuildCacheReq struct {
}

func NewClearBuildCacheReq() *ClearBuildCacheReq {
	return &ClearBuildCacheReq{}
}

func (req *ClearBuildCacheReq) Validate() apperrors.ValidationErrors {
	return nil
}

type ClearBuildCacheResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *ClearBuildCacheDataResp `json:"data"`
}

type ClearBuildCacheDataResp struct {
	CachesDeleted  int    `json:"filesDeleted"`
	SpaceReclaimed uint64 `json:"spaceReclaimed"`
}
