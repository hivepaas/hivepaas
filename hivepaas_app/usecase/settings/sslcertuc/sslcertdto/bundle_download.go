package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DownloadBundleReq struct {
	settings.GetSettingReq
}

func NewDownloadBundleReq() *DownloadBundleReq {
	return &DownloadBundleReq{}
}

func (req *DownloadBundleReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadBundleResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DownloadBundleDataResp `json:"data"`
}

type DownloadBundleDataResp struct {
	*settings.BaseDownloadDataResp
}
