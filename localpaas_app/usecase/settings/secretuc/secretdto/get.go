package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSecretReq struct {
	settings.GetSettingReq
}

func NewGetSecretReq() *GetSecretReq {
	return &GetSecretReq{}
}

func (req *GetSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SecretResp   `json:"data"`
}

type SecretResp struct {
	*settings.BaseSettingResp
	Key      string              `json:"key"`
	Base64   bool                `json:"base64"`
	SwarmRef *SwarmSecretRefResp `json:"swarmRef"`
}

type SwarmSecretRefResp struct {
	File *SwarmRefFileTargetResp `json:"file"`
}

type SwarmRefFileTargetResp struct {
	Name string            `json:"name"`
	UID  string            `json:"uid"`
	GID  string            `json:"gid"`
	Mode fileutil.FileMode `json:"mode"`
}

func TransformSecret(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *SecretResp, err error) {
	secret := setting.MustAsSecret()
	if err = copier.Copy(&resp, &secret); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Key = setting.Name

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
