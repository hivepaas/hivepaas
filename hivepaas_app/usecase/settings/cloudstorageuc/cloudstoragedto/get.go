package cloudstoragedto

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

type GetCloudStorageReq struct {
	settings.GetSettingReq
}

func NewGetCloudStorageReq() *GetCloudStorageReq {
	return &GetCloudStorageReq{}
}

func (req *GetCloudStorageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCloudStorageResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *CloudStorageResp `json:"data"`
}

type CloudStorageResp struct {
	*settings.BaseSettingResp
	S3           *CloudStorageS3Resp `json:"s3"`
	SecretMasked bool                `json:"secretMasked,omitempty"`
}

type CloudStorageS3Resp struct {
	CloudProviderAWSResp
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
	Endpoint string `json:"endpoint"`
}

type CloudProviderAWSResp struct {
	AccessKeyID string `json:"accessKeyId"`
	SecretKey   string `json:"secretKey"`
	Region      string `json:"region"`
}

func (resp *CloudProviderAWSResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

func TransformCloudStorage(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *CloudStorageResp, err error) {
	config := setting.MustAsCloudStorage()
	if err = copier.Copy(&resp, &config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp.SecretMasked = resp.Inherited || (config.S3 != nil && config.S3.SecretKey.IsEncrypted())
	if resp.SecretMasked {
		if resp.S3 != nil {
			resp.S3.SecretKey = maskedSecret
		}
	}

	return resp, nil
}
