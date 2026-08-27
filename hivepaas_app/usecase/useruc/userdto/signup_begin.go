package userdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type BeginUserSignupReq struct {
	InviteToken string `json:"inviteToken"`
}

func NewBeginUserSignupReq() *BeginUserSignupReq {
	return &BeginUserSignupReq{}
}

func (req *BeginUserSignupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.InviteToken, true,
		1, inviteTokenMaxLen, "inviteToken")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginUserSignupResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *BeginUserSignupDataResp `json:"data"`
}

type BeginUserSignupDataResp struct {
	Username       string                  `json:"username"`
	Email          string                  `json:"email"`
	Role           base.UserRole           `json:"role"`
	SecurityOption base.UserSecurityOption `json:"securityOption"`
	AccessExpireAt *time.Time              `json:"accessExpireAt"`

	MFATotpSecret string             `json:"mfaTotpSecret,omitempty"`
	QRCode        *MFATotpQRCodeResp `json:"qrCode,omitempty"`
}
