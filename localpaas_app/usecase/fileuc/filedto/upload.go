package filedto

import (
	"mime/multipart"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UploadReq struct {
	Scope       *base.ObjectScope
	FileType    base.FileType
	StorageType base.FileStorageType
	StorageID   string
	Files       []*multipart.FileHeader
}

func NewUploadReq() *UploadReq {
	return &UploadReq{}
}

func (req *UploadReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(&req.FileType, true, base.AllFileTypes, "type")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UploadResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*FileResp   `json:"data"`
}
