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
	Name string `json:"name"`
	// AccessAction narrows what the key may do below what its owner may do; it
	// never widens anything. It is required: left unset the key carries the
	// owner's full authority, which is the last thing anyone chooses on purpose
	// and the easiest thing to end up with by saying nothing.
	AccessAction *base.AccessActions `json:"accessAction"`
	ExpireAt     time.Time           `json:"expireAt"`
	// Capabilities are those of the owner's the key may use - revealing secrets.
	// None by default: a key given to a script or an assistant should not hand
	// it passwords and keys in the clear unless somebody chose that.
	Capabilities []base.ResourceCapability `json:"capabilities"`
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
	// A key granting nothing is refused as well as one granting everything by
	// omission: it cannot do any work, so it is a mistake either way.
	validators = append(validators, basedto.ValidateCond(
		req.AccessAction != nil && !req.AccessAction.IsNoAccess(), "accessAction")...)
	validators = append(validators, basedto.ValidateSlice(req.Capabilities, true, 0,
		base.APIKeyCapabilities, "capabilities")...)
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
