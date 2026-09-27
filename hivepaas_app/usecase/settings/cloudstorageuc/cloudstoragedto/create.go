package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maxKeyLen = 100
)

type CreateCloudStorageReq struct {
	settings.CreateSettingReq
	*CloudStorageBaseReq
}

type CloudStorageBaseReq struct {
	Name string                `json:"name"`
	Kind base.CloudStorageKind `json:"kind"`
	S3   *CloudStorageS3Req    `json:"s3"`
}

func (req *CloudStorageBaseReq) ToEntity() *entity.CloudStorage {
	res := &entity.CloudStorage{}
	if req.Kind == base.CloudStorageKindS3 {
		res.S3 = req.S3.ToEntity()
	}
	return res
}

type CloudStorageS3Req struct {
	// KeyAuth is the key auth setting the bucket is reached with.
	KeyAuth  basedto.ObjectIDReq `json:"keyAuth"`
	Region   string              `json:"region"`
	Bucket   string              `json:"bucket"`
	Endpoint string              `json:"endpoint"`
}

func (req *CloudStorageS3Req) ToEntity() *entity.CloudStorageS3 {
	if req == nil {
		return nil
	}
	return &entity.CloudStorageS3{
		KeyAuth:  entity.ObjectID{ID: req.KeyAuth.ID},
		Region:   req.Region,
		Bucket:   req.Bucket,
		Endpoint: req.Endpoint,
	}
}

func (req *CloudStorageS3Req) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateObjectIDReq(&req.KeyAuth, true, field+"keyAuth")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 1, maxKeyLen, field+"region")...)
	res = append(res, basedto.ValidateStr(&req.Bucket, true, 1, maxKeyLen, field+"bucket")...)
	res = append(res, basedto.ValidateStr(&req.Endpoint, false, 1, maxKeyLen, field+"endpoint")...)
	return res
}

func (req *CloudStorageBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllCloudStorageKinds, field+"kind")...)
	if req.Kind == base.CloudStorageKindS3 {
		res = append(res, req.S3.validate("s3")...)
	}
	return res
}

func NewCreateCloudStorageReq() *CreateCloudStorageReq {
	return &CreateCloudStorageReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCloudStorageReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCloudStorageResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
