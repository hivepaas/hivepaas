package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	findKindMaxLen = 100
	findPathMaxLen = 200
	bucketMaxLen   = 100
)

type CreateFileReq struct {
	Scope *base.ObjectScope `json:"-"`

	FileType  base.FileType `json:"fileType"`
	FileKind  base.FileKind `json:"fileKind"`
	FilePath  string        `json:"filePath"`
	StorageID string        `json:"storageId"`
	Bucket    string        `json:"bucket"`
}

func NewCreateFileReq() *CreateFileReq {
	return &CreateFileReq{}
}

func (req *CreateFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(&req.FileType, true, base.AllFileTypes,
		"fileType")...)
	validators = append(validators, basedto.ValidateStr(&req.FileKind, true, 1, findKindMaxLen,
		"fileKind")...)
	validators = append(validators, basedto.ValidateStr(&req.FilePath, true, 1, findPathMaxLen,
		"findPath")...)
	validators = append(validators, basedto.ValidateID(&req.StorageID, true, "storageId")...)
	validators = append(validators, basedto.ValidateStr(&req.Bucket, true, 1, bucketMaxLen,
		"bucket")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateFileResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
