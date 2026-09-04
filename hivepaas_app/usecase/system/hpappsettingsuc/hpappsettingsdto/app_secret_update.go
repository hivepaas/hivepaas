package hpappsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
)

const (
	secretMaxLen = secrethelper.DefaultPasswordMaxLen
)

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
		1, secretMaxLen, "currentSecret")...)
	validators = append(validators, basedto.ValidateStr(&req.NewSecret, true,
		1, secretMaxLen, "newSecret")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
}
