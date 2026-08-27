package appdto

import (
	"encoding/base64"
	"path/filepath"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	projectPhotoMaxSize = 300 * 1024 // 300KB
)

type UpdateAppPhotoReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	*AppPhotoReq
}

type AppPhotoReq struct {
	Delete       bool   `json:"delete"`
	IsPresetIcon bool   `json:"isPresetIcon"`
	FileName     string `json:"fileName"`
	DataBase64   string `json:"dataBase64"`

	// NOTE: Use locally only
	DataBytes []byte `json:"-"`
}

func (req *AppPhotoReq) IsChanged() bool {
	if req == nil {
		return false
	}
	return req.Delete || req.FileName != ""
}

func (req *AppPhotoReq) modifyRequest() error {
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

func (req *AppPhotoReq) validate(field string) []vld.Validator {
	if req == nil || req.FileName == "" {
		return nil
	}
	if req.IsPresetIcon {
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

func NewUpdateAppPhotoReq() *UpdateAppPhotoReq {
	return &UpdateAppPhotoReq{}
}

func (req *UpdateAppPhotoReq) ModifyRequest() error {
	return req.modifyRequest()
}

func (req *UpdateAppPhotoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppPhotoResp struct {
	Meta *basedto.Meta `json:"meta"`
}
