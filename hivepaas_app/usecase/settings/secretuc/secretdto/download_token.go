package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetDownloadTokenReq struct {
	settings.GetSettingReq
	DataType   string            `json:"-" mapstructure:"-"`
	Expiration timeutil.Duration `json:"-" mapstructure:"-"`
}

func NewGetDownloadTokenReq() *GetDownloadTokenReq {
	return &GetDownloadTokenReq{}
}

func (req *GetDownloadTokenReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDownloadTokenResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *GetDownloadTokenDataResp `json:"data"`
}

type GetDownloadTokenDataResp struct {
	Token string `json:"token"`
}
