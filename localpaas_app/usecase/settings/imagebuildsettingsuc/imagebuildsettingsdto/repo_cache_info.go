package imagebuildsettingsdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
)

type GetRepoCacheInfoReq struct {
	Scope *base.ObjectScope `json:"-" mapstructure:"-"`
}

func NewGetRepoCacheInfoReq() *GetRepoCacheInfoReq {
	return &GetRepoCacheInfoReq{}
}

func (req *GetRepoCacheInfoReq) Validate() apperrors.ValidationErrors {
	// TODO: add validation
	return nil
}

type GetRepoCacheInfoResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *RepoCacheInfoResp `json:"data"`
}

type RepoCacheInfoResp struct {
	TotalFiles  int             `json:"totalFiles"`
	TotalSizeHR unit.DataSizeHR `json:"totalSizeHR"` // human-readable
}

func TransformRepoCacheInfo(files []*entity.File) (resp *RepoCacheInfoResp) {
	resp = &RepoCacheInfoResp{}
	var totalSize int64
	for _, file := range files {
		if file.Deleted {
			continue
		}
		resp.TotalFiles++
		totalSize += file.Size
	}
	resp.TotalSizeHR = unit.DataSizeHR(totalSize)
	return resp
}
