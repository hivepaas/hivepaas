package cloudproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetCloudProviderReq struct {
	settings.GetSettingReq
}

func NewGetCloudProviderReq() *GetCloudProviderReq {
	return &GetCloudProviderReq{}
}

func (req *GetCloudProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCloudProviderResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *CloudProviderResp `json:"data"`
}

type CloudProviderResp struct {
	*settings.BaseSettingResp
	AWS          *CloudProviderAWSResp `json:"aws,omitempty"`
	SecretMasked bool                  `json:"secretMasked,omitempty"`
}

type CloudProviderAWSResp struct {
	AccessKeyID string `json:"accessKeyID"`
	SecretKey   string `json:"secretKey,omitempty"`
	Region      string `json:"region"`
}

func (resp *CloudProviderAWSResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

func TransformCloudProvider(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *CloudProviderResp, err error) {
	if setting == nil {
		return nil, nil
	}

	config := setting.MustAsCloudProvider()
	if err = copier.Copy(&resp, &config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.SecretMasked = resp.Inherited || (config.AWS != nil && config.AWS.SecretKey.IsEncrypted())
	if resp.SecretMasked {
		if config.AWS != nil {
			resp.AWS.SecretKey = maskedSecret
		}
	}

	return resp, nil
}
