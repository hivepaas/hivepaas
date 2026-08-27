package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DownloadBundleReq struct {
	settings.GetSettingReq
}

func NewDownloadBundleReq() *DownloadBundleReq {
	return &DownloadBundleReq{}
}

func (req *DownloadBundleReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadBundleResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DownloadBundleDataResp `json:"data"`
}

type DownloadBundleDataResp struct {
	*settings.BaseDownloadDataResp
}
