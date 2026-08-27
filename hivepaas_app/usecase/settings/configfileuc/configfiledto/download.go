package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DownloadConfigFileReq struct {
	settings.GetSettingReq
	DataType   string `json:"-" mapstructure:"-"`
	Token      string `json:"-" mapstructure:"token"`
	ViewInline bool   `json:"-" mapstructure:"viewInline"`
}

func NewDownloadConfigFileReq() *DownloadConfigFileReq {
	return &DownloadConfigFileReq{}
}

func (req *DownloadConfigFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadConfigFileResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *DownloadConfigFileDataResp `json:"data"`
}

type DownloadConfigFileDataResp struct {
	*settings.BaseDownloadDataResp
}
