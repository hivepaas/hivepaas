package apikeydto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
func (req *CreateAPIKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	timeNow := timeutil.NowUTC()
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1,
		base.SettingNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateTime(&req.ExpireAt, false, timeNow,
		timeNow.AddDate(expirationYearMax, 0, 0), "expireAt")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAPIKeyResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *APIKeyDataResp `json:"data"`
}

type APIKeyDataResp struct {
	ID        string `json:"id"`
	KeyID     string `json:"keyID"`
	SecretKey string `json:"secretKey"`
}
