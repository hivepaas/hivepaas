package sessiondto

import (
	"github.com/markbates/goth"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type CreateOAuthSessionReq struct {
	User *goth.User
}

func NewCreateOAuthSessionReq() *CreateOAuthSessionReq {
	return &CreateOAuthSessionReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateOAuthSessionReq) Validate() hperrors.ValidationErrors {
	return nil
}

type CreateOAuthSessionResp struct {
	BaseCreateSessionResp
}
