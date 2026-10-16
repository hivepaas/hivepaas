package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RefreshSessionReq struct{}

func NewRefreshSessionReq() *RefreshSessionReq {
	return &RefreshSessionReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RefreshSessionReq) Validate() hperrors.ValidationErrors {
	return nil
}

type RefreshSessionResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *RefreshSessionDataResp `json:"data"`
}

type RefreshSessionDataResp struct {
	*BaseCreateSessionResp
}
