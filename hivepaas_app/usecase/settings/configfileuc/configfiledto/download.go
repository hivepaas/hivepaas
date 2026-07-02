package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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

func (req *DownloadConfigFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadConfigFileResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *DownloadConfigFileDataResp `json:"data"`
}

type DownloadConfigFileDataResp struct {
	*settings.BaseDownloadDataResp
}
