package secretdto

import (
	"os"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type ListSecretReq struct {
	settings.ListSettingReq
}

func NewListSecretReq() *ListSecretReq {
	return &ListSecretReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSecretResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SecretResp     `json:"data"`
}

type SecretResp struct {
	*settings.BaseSettingResp
	Key      string              `json:"key"`
	Base64   bool                `json:"base64"`
	SwarmRef *SwarmSecretRefResp `json:"swarmRef"`
}

type SwarmSecretRefResp struct {
	File *SwarmSecretRefFileTargetResp `json:"file"`
}

type SwarmSecretRefFileTargetResp struct {
	Name string      `json:"name"`
	UID  string      `json:"uid"`
	GID  string      `json:"gid"`
	Mode os.FileMode `json:"mode"`
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

func TransformSecrets(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*SecretResp, err error) {
	resp = make([]*SecretResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSecret(setting, refObjects)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
