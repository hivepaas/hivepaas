package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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

func (req *DownloadSecretReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadSecretResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DownloadSecretDataResp `json:"data"`
}

type DownloadSecretDataResp struct {
	*settings.BaseDownloadDataResp
}
