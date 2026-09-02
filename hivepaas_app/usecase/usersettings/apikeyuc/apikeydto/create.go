package apikeydto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	expirationYearMax = 1
)

type CreateAPIKeyReq struct {
	settings.CreateSettingReq
	Name         string              `json:"name"`
	AccessAction *base.AccessActions `json:"accessAction"`
	ExpireAt     time.Time           `json:"expireAt"`
}

func NewCreateAPIKeyReq() *CreateAPIKeyReq {
	return &CreateAPIKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAPIKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	timeNow := timeutil.NowUTC()
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1,
		base.SettingNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateTime(&req.ExpireAt, true, timeNow,
		timeNow.AddDate(expirationYearMax, 0, 0), "expireAt")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAPIKeyResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *APIKeyDataResp `json:"data"`
}

type APIKeyDataResp struct {
	ID        string `json:"id"`
	KeyID     string `json:"keyId"`
	SecretKey string `json:"secretKey"`
}
