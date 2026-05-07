package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetConfigFileReq struct {
	settings.GetSettingReq
}

func NewGetConfigFileReq() *GetConfigFileReq {
	return &GetConfigFileReq{}
}

func (req *GetConfigFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetConfigFileResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *ConfigFileResp `json:"data"`
}

type ConfigFileResp struct {
	*settings.BaseSettingResp
	Content  string                  `json:"content"`
	Base64   bool                    `json:"base64"`
	SwarmRef *SwarmConfigFileRefResp `json:"swarmRef"`
}

type SwarmConfigFileRefResp struct {
	File *SwarmRefFileTargetResp `json:"file"`
}

type SwarmRefFileTargetResp struct {
	Name string            `json:"name"`
	UID  string            `json:"uid"`
	GID  string            `json:"gid"`
	Mode fileutil.FileMode `json:"mode"`
}

func TransformConfigFile(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *ConfigFileResp, err error) {
	configFile := setting.MustAsConfigFile()
	if err = copier.Copy(&resp, &configFile); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Name = setting.Name

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
