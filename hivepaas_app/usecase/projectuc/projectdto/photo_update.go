package projectdto

import (
	"encoding/base64"
	"path/filepath"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	projectPhotoMaxSize = 300 * 1024 // 300KB
)

type UpdateProjectPhotoReq struct {
	ID string `json:"id"`
	*ProjectPhotoReq
}

type ProjectPhotoReq struct {
	Delete     bool   `json:"delete"`
	FileName   string `json:"fileName"`
	DataBase64 string `json:"dataBase64"`

	// NOTE: Use locally only
	DataBytes []byte `json:"-"`
}

func (req *ProjectPhotoReq) IsChanged() bool {
	if req == nil {
		return false
	}
	return req.Delete || req.FileName != ""
}

func (req *ProjectPhotoReq) modifyRequest() error {
	if req != nil && req.DataBase64 != "" {
		dataBase64 := req.DataBase64
		// Image base64 from FE can be in form: `data:image/png;base64,<data-in-base64>`
		if strings.HasPrefix(dataBase64, "data:") {
			dataBase64 = dataBase64[strings.Index(dataBase64, ",")+1:]
		}
		req.DataBytes, _ = base64.StdEncoding.DecodeString(dataBase64)
	}
	return nil
}

func (req *ProjectPhotoReq) validate(field string) []vld.Validator {
	if req == nil || req.FileName == "" {
		return nil
	}
	if field != "" {
		field += "."
	}
	fileExt := strings.ToLower(filepath.Ext(req.FileName))
	return []vld.Validator{
		vld.Must(gofn.Contain(base.AllPhotoFileExts, fileExt)).OnError(
			vld.SetField(field+"fileName", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_EXT_UNSUPPORTED"),
		),
		vld.Must(len(req.DataBytes) > 0).OnError(
			vld.SetField(field+"dataBase64", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_INVALID"),
		),
		vld.When(len(req.DataBytes) > 0).Then(
			vld.Must(len(req.DataBytes) <= projectPhotoMaxSize).OnError(
				vld.SetField(field+"dataBase64", nil),
				vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_TOO_BIG"),
			),
		),
	}
}

func NewUpdateProjectPhotoReq() *UpdateProjectPhotoReq {
	return &UpdateProjectPhotoReq{}
}

func (req *UpdateProjectPhotoReq) ModifyRequest() error {
	return req.modifyRequest()
}

func (req *UpdateProjectPhotoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectPhotoResp struct {
	Meta *basedto.Meta `json:"meta"`
}
