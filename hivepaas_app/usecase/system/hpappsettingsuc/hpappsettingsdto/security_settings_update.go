package hpappsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateSecuritySettingsReq struct {
	// AppSecret is the current app secret, re-entered. It is not authentication -
	// the endpoint is already admin-only - it is proof that the caller is the
	// operator and not merely a session that reached an admin account. The flags
	// below decide whether stored secrets may leave the server, so a stolen
	// session must not be enough to turn them on.
	AppSecret string `json:"appSecret"`

	ReturnSecretsViaAPI bool `json:"returnSecretsViaApi"`
}

func NewUpdateSecuritySettingsReq() *UpdateSecuritySettingsReq {
	return &UpdateSecuritySettingsReq{}
}

// ToConfig builds the security settings this request asks for. Every field is
// carried, present or not: the request is the whole new state, so a flag the
// caller left at false is a flag they want off.
func (req *UpdateSecuritySettingsReq) ToConfig() *config.Security {
	return &config.Security{
		ReturnSecretsViaAPI: req.ReturnSecretsViaAPI,
	}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSecuritySettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.AppSecret, true,
		1, existingAppSecretMaxLen, "appSecret")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSecuritySettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
