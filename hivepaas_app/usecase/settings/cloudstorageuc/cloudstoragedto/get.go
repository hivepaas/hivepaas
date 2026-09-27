package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetCloudStorageReq struct {
	settings.GetSettingReq
}

func NewGetCloudStorageReq() *GetCloudStorageReq {
	return &GetCloudStorageReq{}
}

func (req *GetCloudStorageReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCloudStorageResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *CloudStorageResp `json:"data"`
}

type CloudStorageResp struct {
	*settings.BaseSettingResp
	S3 *CloudStorageS3Resp `json:"s3"`
}

type CloudStorageS3Resp struct {
	// KeyAuth is the key auth the bucket is reached with: its id and name, or
	// its id alone and missing when it no longer exists.
	KeyAuth  *settings.BaseSettingResp `json:"keyAuth"`
	Region   string                    `json:"region"`
	Bucket   string                    `json:"bucket"`
	Endpoint string                    `json:"endpoint"`
}

func TransformCloudStorage(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *CloudStorageResp, err error) {
	config := setting.MustAsCloudStorage()
	resp = &CloudStorageResp{}
	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if config.S3 != nil {
		resp.S3 = &CloudStorageS3Resp{Region: config.S3.Region, Bucket: config.S3.Bucket,
			Endpoint: config.S3.Endpoint}
		resp.S3.KeyAuth, err = keyAuthOf(config.S3.KeyAuth.ID, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return resp, nil
}

// keyAuthOf is the linked key auth as a reference: never its key, which is the
// key auth's own to reveal.
func keyAuthOf(id string, refObjects *entity.RefObjects) (*settings.BaseSettingResp, error) {
	if id == "" {
		return nil, nil
	}
	var keyAuth *entity.Setting
	if refObjects != nil {
		keyAuth = refObjects.RefSettings[id]
	}
	if keyAuth == nil {
		return settings.NewMissingSetting(id, base.SettingTypeKeyAuth), nil
	}
	resp, err := settings.TransformSettingBase(keyAuth)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
