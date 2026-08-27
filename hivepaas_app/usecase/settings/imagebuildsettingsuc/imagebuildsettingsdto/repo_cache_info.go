package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetRepoCacheInfoReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewGetRepoCacheInfoReq() *GetRepoCacheInfoReq {
	return &GetRepoCacheInfoReq{}
}

func (req *GetRepoCacheInfoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateCond(req.Scope.IsValid(), "params")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRepoCacheInfoResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *RepoCacheInfoResp `json:"data"`
}

type RepoCacheInfoResp struct {
	TotalFiles     int   `json:"totalFiles"`
	TotalSizeBytes int64 `json:"totalSizeBytes"`
}

func TransformRepoCacheInfo(files []*entity.File) (resp *RepoCacheInfoResp) {
	resp = &RepoCacheInfoResp{}
	resp.TotalSizeBytes = 0
	for _, file := range files {
		if file.Deleted {
			continue
		}
		resp.TotalFiles++
		resp.TotalSizeBytes += file.Size
	}
	return resp
}
