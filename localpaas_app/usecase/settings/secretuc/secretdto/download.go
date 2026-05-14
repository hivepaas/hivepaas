package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type DownloadSecretReq struct {
	settings.GetSettingReq
	DataType   string `json:"-" mapstructure:"-"`
	Token      string `json:"-" mapstructure:"token"`
	ViewInline bool   `json:"-" mapstructure:"viewInline"`
}

func NewDownloadSecretReq() *DownloadSecretReq {
	return &DownloadSecretReq{}
}

func (req *DownloadSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadSecretResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DownloadSecretDataResp `json:"data"`
}

type DownloadSecretDataResp struct {
	*settings.BaseDownloadDataResp
}
