package hpappsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

// existingAppSecretMaxLen bounds the secret the caller submits to prove they are
// the operator.
//
// It is deliberately not hpappservice.SecretRequirements.MaxLen. That bound
// governs a secret we are about to impose and can therefore keep short; this one
// only has to accommodate a secret already in use, and a development install
// generates itself a 64-character one - which the stricter bound would reject,
// locking that install out of its own settings.
const existingAppSecretMaxLen = 200

type UpdateAppSecretReq struct {
	CurrentSecret string `json:"currentSecret"`
	NewSecret     string `json:"newSecret"`
}

func NewUpdateAppSecretReq() *UpdateAppSecretReq {
	return &UpdateAppSecretReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppSecretReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.CurrentSecret, true,
		1, existingAppSecretMaxLen, "currentSecret")...)
	validators = append(validators, basedto.ValidateStr(&req.NewSecret, true,
		hpappservice.SecretRequirements.MinLen, hpappservice.SecretRequirements.MaxLen, "newSecret")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
}
