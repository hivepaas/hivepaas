package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type GetRepoCacheInfoReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewGetRepoCacheInfoReq() *GetRepoCacheInfoReq {
	return &GetRepoCacheInfoReq{}
}

func (req *GetRepoCacheInfoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateCond(req.Scope.IsValid(), "params")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
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
