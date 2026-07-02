package sessiondto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

const (
	maxUsernameLen = 100
	maxEmailLen    = 100
	maxPasswordLen = 100
)

type LoginWithPasswordReq struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	TrustedDeviceID string `json:"trustedDeviceId"`

	AcceptLanguage translation.Lang `json:"-"`
}

func NewLoginWithPasswordReq() *LoginWithPasswordReq {
	return &LoginWithPasswordReq{}
}

func (req *LoginWithPasswordReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Username = strings.TrimSpace(req.Username)
	req.TrustedDeviceID = strings.TrimSpace(req.TrustedDeviceID)
	return nil
}

func (req *LoginWithPasswordReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Username, true, 1,
		maxUsernameLen, "username")...)
	validators = append(validators, basedto.ValidateStr(&req.Password, true, 1,
		maxPasswordLen, "password")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginWithPasswordResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *LoginWithPasswordDataResp `json:"data"`
}

type LoginWithPasswordDataResp struct {
	NextStep string                 `json:"nextStep,omitempty"`
	MFAType  base.MFAType           `json:"mfaType,omitempty"`
	MFAToken string                 `json:"mfaToken,omitempty"`
	Session  *BaseCreateSessionResp `json:"session,omitempty"`
}
