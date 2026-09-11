package hpappsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetSecuritySettingsReq struct {
}

func NewGetSecuritySettingsReq() *GetSecuritySettingsReq {
	return &GetSecuritySettingsReq{}
}

func (req *GetSecuritySettingsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetSecuritySettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *SecuritySettingsResp `json:"data"`
}

type SecuritySettingsResp struct {
	ReturnSecretsViaAPI     bool     `json:"returnSecretsViaApi"`
	AlwaysReturnSecretTypes []string `json:"alwaysReturnSecretTypes"`
}

type SecuritySettingsTransformInput struct {
	Config *config.Config
}

func TransformSecuritySettings(input *SecuritySettingsTransformInput) (resp *SecuritySettingsResp, err error) {
	// Normalised to an empty list rather than passed through as nil: the field is
	// a set the client renders, and "no exemptions" arriving as null makes every
	// reader handle a case that carries no extra meaning.
	exemptions := input.Config.Security.AlwaysReturnSecretTypes
	if exemptions == nil {
		exemptions = []string{}
	}
	resp = &SecuritySettingsResp{
		ReturnSecretsViaAPI:     input.Config.Security.ReturnSecretsViaAPI,
		AlwaysReturnSecretTypes: exemptions,
	}
	return resp, nil
}
