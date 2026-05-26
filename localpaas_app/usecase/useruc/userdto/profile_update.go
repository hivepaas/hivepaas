package userdto

import (
	"encoding/base64"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
)

type UpdateProfileReq struct {
	Username string        `json:"username"`
	Email    string        `json:"email"`
	FullName string        `json:"fullName"`
	Position *string       `json:"position"`
	Photo    *UserPhotoReq `json:"photo"`
	Notes    *string       `json:"notes"`
}

func NewUpdateProfileReq() *UpdateProfileReq {
	return &UpdateProfileReq{}
}

func (req *UpdateProfileReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strutil.NormalizeEmail(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	// Parse photo
	if req.Photo != nil && req.Photo.DataBase64 != "" {
		dataBase64 := req.Photo.DataBase64
		// Image base64 from FE can be in form: `data:image/png;base64,<data-in-base64>`
		if strings.HasPrefix(dataBase64, "data:") {
			dataBase64 = dataBase64[strings.Index(dataBase64, ",")+1:]
		}
		req.Photo.DataBytes, _ = base64.StdEncoding.DecodeString(dataBase64)
	}
	return nil
}

func (req *UpdateProfileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, validateUsername(&req.Username, false, "username")...)
	validators = append(validators, basedto.ValidateEmail(&req.Email, false, "email")...)
	validators = append(validators, basedto.ValidateStr(&req.FullName, false,
		nameMinLen, nameMaxLen, "fullName")...)
	validators = append(validators, basedto.ValidateStr(req.Position, false,
		nameMinLen, nameMaxLen, "position")...)
	validators = append(validators, validateUserPhoto(req.Photo, "photo")...)
	validators = append(validators, basedto.ValidateStr(req.Notes, false,
		notesMinLen, notesMaxLen, "notes")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProfileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
