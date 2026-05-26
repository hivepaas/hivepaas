package projectdto

import (
	"encoding/base64"
	"path/filepath"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	maxProjectPhotoSize = 300 * 1024 // 300KB
)

type UpdateProjectPhotoReq struct {
	ID         string `json:"id"`
	Delete     bool   `json:"delete"`
	FileName   string `json:"fileName"`
	DataBase64 string `json:"dataBase64"`

	// NOTE: Use locally only
	DataBytes []byte `json:"-"`
}

func NewUpdateProjectPhotoReq() *UpdateProjectPhotoReq {
	return &UpdateProjectPhotoReq{}
}

func (req *UpdateProjectPhotoReq) ModifyRequest() error {
	if req.FileName != "" && req.DataBase64 != "" {
		req.DataBytes, _ = base64.StdEncoding.DecodeString(req.DataBase64)
	}
	return nil
}

func validateProjectPhoto(photo *UpdateProjectPhotoReq) []vld.Validator {
	if photo == nil || photo.FileName == "" {
		return nil
	}
	fileExt := strings.ToLower(filepath.Ext(photo.FileName))
	return []vld.Validator{
		vld.Must(gofn.Contain(base.AllPhotoFileExts, fileExt)).OnError(
			vld.SetField("fileName", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_EXT_UNSUPPORTED"),
		),
		vld.Must(len(photo.DataBytes) > 0).OnError(
			vld.SetField("dataBase64", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_INVALID"),
		),
		vld.When(len(photo.DataBytes) > 0).Then(
			vld.Must(len(photo.DataBytes) <= maxProjectPhotoSize).OnError(
				vld.SetField("dataBase64", nil),
				vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_TOO_BIG"),
			),
		),
	}
}

func (req *UpdateProjectPhotoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, validateProjectPhoto(req)...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectPhotoResp struct {
	Meta *basedto.Meta `json:"meta"`
}
