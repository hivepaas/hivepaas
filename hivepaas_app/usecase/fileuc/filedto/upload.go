package filedto

import (
	"mime/multipart"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type UploadReq struct {
	Scope       *entity.ObjectScope
	FileType    base.FileType
	FileKind    base.FileKind
	StorageType base.FileStorageType
	StorageID   string
	Files       []*multipart.FileHeader
}

func NewUploadReq() *UploadReq {
	return &UploadReq{}
}

func (req *UploadReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStrIn(&req.FileType, true, base.AllFileTypes, "type")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UploadResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*FileResp   `json:"data"`
}
