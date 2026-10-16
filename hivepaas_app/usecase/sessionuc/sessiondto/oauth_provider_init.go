package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type InitOAuthProviderReq struct {
	// Provider is the ID of the OAuth setting signed in with, from the path.
	Provider string `json:"-" mapstructure:"-"`
}

func NewInitOAuthProviderReq() *InitOAuthProviderReq {
	return &InitOAuthProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *InitOAuthProviderReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.Provider, true, "provider")...))
}
