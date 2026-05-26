package userdto

import (
	"encoding/base64"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
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

func NewCompleteUserSignupReq() *CompleteUserSignupReq {
	return &CompleteUserSignupReq{}
}

func (req *CompleteUserSignupReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.FullName = strings.TrimSpace(req.FullName)
	// Parse photo
	if req.Photo != nil && req.Photo.DataBase64 != "" {
		req.Photo.DataBytes, _ = base64.StdEncoding.DecodeString(req.Photo.DataBase64)
	}
	return nil
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
