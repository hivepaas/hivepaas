package builddto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
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
