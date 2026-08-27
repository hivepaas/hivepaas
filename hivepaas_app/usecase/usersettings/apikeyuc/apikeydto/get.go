package apikeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetAPIKeyReq struct {
	settings.GetSettingReq
}

func NewGetAPIKeyReq() *GetAPIKeyReq {
	return &GetAPIKeyReq{}
}

func (req *GetAPIKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAPIKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *APIKeyResp   `json:"data"`
}

type APIKeyResp struct {
	*settings.BaseSettingResp
	KeyID        string             `json:"keyId"`
	AccessAction base.AccessActions `json:"accessAction"`
}

func TransformAPIKey(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *APIKeyResp, err error) {
	apiKey := setting.MustAsAPIKey()
	if err = copier.Copy(&resp, apiKey); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
