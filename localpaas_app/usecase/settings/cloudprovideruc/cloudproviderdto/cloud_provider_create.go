package cloudproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	maxKeyLen = 255
)

type CreateCloudProviderReq struct {
	settings.CreateSettingReq
	*CloudProviderBaseReq
}

type CloudProviderBaseReq struct {
	Name string                 `json:"name"`
	Kind base.CloudProviderKind `json:"kind"`
	AWS  *CloudProviderAWSReq   `json:"aws"`
}

func (req *CloudProviderBaseReq) ToEntity() *entity.CloudProvider {
	res := &entity.CloudProvider{}
	if req.Kind == base.CloudProviderKindAWS {
		res.AWS = req.AWS.ToEntity()
	}
	return res
}

type CloudProviderAWSReq struct {
	AccessKeyID string `json:"accessKeyID"`
	SecretKey   string `json:"secretKey"`
	Region      string `json:"region"`
}

func (req *CloudProviderAWSReq) ToEntity() *entity.CloudProviderAWS {
	if req == nil {
		return nil
	}
	return &entity.CloudProviderAWS{
		AccessKeyID: req.AccessKeyID,
		SecretKey:   entity.NewEncryptedField(req.SecretKey),
		Region:      req.Region,
	}
}

func (req *CloudProviderAWSReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AccessKeyID, true, 1, maxKeyLen, field+"accessKeyID")...)
	res = append(res, basedto.ValidateStr(&req.SecretKey, true, 1, maxKeyLen, field+"secretKey")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 1, maxKeyLen, field+"region")...)
	return res
}

func (req *CloudProviderBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllCloudProviderKinds, field+"kind")...)
	if req.Kind == base.CloudProviderKindAWS {
		res = append(res, req.AWS.validate("")...)
	}
	return res
}

func NewCreateCloudProviderReq() *CreateCloudProviderReq {
	return &CreateCloudProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCloudProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCloudProviderResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
