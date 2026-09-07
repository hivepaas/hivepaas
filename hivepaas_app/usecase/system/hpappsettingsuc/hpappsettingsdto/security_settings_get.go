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
	ReturnSecretsViaAPI bool `json:"returnSecretsViaApi"`
}

type SecuritySettingsTransformInput struct {
	Config *config.Config
}

func TransformSecuritySettings(input *SecuritySettingsTransformInput) (resp *SecuritySettingsResp, err error) {
	resp = &SecuritySettingsResp{
		ReturnSecretsViaAPI: input.Config.Security.ReturnSecretsViaAPI,
	}
	return resp, nil
}
