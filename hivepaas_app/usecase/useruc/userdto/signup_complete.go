package userdto

import (
	"encoding/base64"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	inviteTokenMaxLen = 10000
)

type CompleteUserSignupReq struct {
	InviteToken string `json:"inviteToken"`

	Username string        `json:"username"`
	FullName string        `json:"fullName"`
	Photo    *UserPhotoReq `json:"photo"`

	// Required when security option is password-2fa/password-only
	Password string `json:"password"`

	// Required when security option is password-2fa
	Passcode      string `json:"passcode"`
	MFATotpSecret string `json:"mfaTotpSecret"`
}

type UserPhotoReq struct {
	Delete     bool   `json:"delete"`
	FileName   string `json:"fileName"`
	DataBase64 string `json:"dataBase64"`

	// NOTE: Use locally only
	DataBytes []byte `json:"-"`
}

func (req *UserPhotoReq) IsChanged() bool {
	if req == nil {
		return false
	}
	return req.Delete || req.FileName != ""
}

func (req *UserPhotoReq) modifyRequest() error {
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

func NewCompleteUserSignupReq() *CompleteUserSignupReq {
	return &CompleteUserSignupReq{}
}

func (req *CompleteUserSignupReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.FullName = strings.TrimSpace(req.FullName)
	return req.Photo.modifyRequest()
}

func (req *CompleteUserSignupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.InviteToken, true,
		1, inviteTokenMaxLen, "inviteToken")...)
	validators = append(validators, validateUsername(&req.Username, true, "username")...)
	validators = append(validators, basedto.ValidateStr(&req.FullName, true,
		nameMinLen, nameMaxLen, "fullName")...)
	validators = append(validators, validateUserPhoto(req.Photo, "photo")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CompleteUserSignupResp struct {
	Meta *basedto.Meta `json:"meta"`
}
