package keyauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetKeyAuthReq struct {
	settings.GetSettingReq
}

func NewGetKeyAuthReq() *GetKeyAuthReq {
	return &GetKeyAuthReq{}
}

func (req *GetKeyAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetKeyAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *KeyAuthResp  `json:"data"`
}

type KeyAuthResp struct {
	*settings.BaseSettingResp
	KeyID        string `json:"keyId"`
	SecretKey    string `json:"secretKey"`
	SecretMasked bool   `json:"secretMasked,omitempty"`
}

func (resp *KeyAuthResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

func TransformKeyAuth(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *KeyAuthResp, err error) {
	config := setting.MustAsKeyAuth()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.SecretMasked = config.SecretKey.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.SecretKey = basedto.MaskedSecret
	}

	return resp, nil
}
