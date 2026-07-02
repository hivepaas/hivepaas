package accesstokendto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetAccessTokenReq struct {
	settings.GetSettingReq
}

func NewGetAccessTokenReq() *GetAccessTokenReq {
	return &GetAccessTokenReq{}
}

func (req *GetAccessTokenReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAccessTokenResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *AccessTokenResp `json:"data"`
}

type AccessTokenResp struct {
	*settings.BaseSettingResp
	User         string `json:"user"`
	Token        string `json:"token"`
	BaseURL      string `json:"baseURL"`
	SecretMasked bool   `json:"secretMasked,omitempty"`
}

func (resp *AccessTokenResp) CopyToken(field entity.EncryptedField) error {
	resp.Token = field.String()
	return nil
}

func TransformAccessToken(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *AccessTokenResp, err error) {
	config := setting.MustAsAccessToken()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp.SecretMasked = config.Token.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.Token = maskedSecret
	}

	return resp, nil
}
