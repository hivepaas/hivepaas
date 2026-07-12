package registryauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "********"
)

type GetRegistryAuthReq struct {
	settings.GetSettingReq
}

func NewGetRegistryAuthReq() *GetRegistryAuthReq {
	return &GetRegistryAuthReq{}
}

func (req *GetRegistryAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRegistryAuthResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *RegistryAuthResp `json:"data"`
}

type RegistryAuthResp struct {
	*settings.BaseSettingResp
	Address      string `json:"address"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Readonly     bool   `json:"readonly"`
	SecretMasked bool   `json:"secretMasked,omitempty"`
}

func (resp *RegistryAuthResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

func TransformRegistryAuth(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *RegistryAuthResp, err error) {
	config := setting.MustAsRegistryAuth()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.SecretMasked = config.Password.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.Password = maskedSecret
	}

	return resp, nil
}
