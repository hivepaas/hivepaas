package binobjectdto

import (
	"io"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetBinObjectDataReq struct {
	ID   string             `json:"-"`
	Type base.BinObjectType `json:"-"`
}

func NewGetBinObjectDataReq() *GetBinObjectDataReq {
	return &GetBinObjectDataReq{}
}

func (req *GetBinObjectDataReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetBinObjectDataResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *BinObjectDataResp `json:"data"`
}

type BinObjectDataResp struct {
	ContentLength int64
	ContentType   string
	Content       io.ReadCloser
	ExtraHeaders  map[string]string
}
