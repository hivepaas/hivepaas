package networkdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type SyncNetworkReq struct {
}

func NewSyncNetworkReq() *SyncNetworkReq {
	return &SyncNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SyncNetworkReq) Validate() apperrors.ValidationErrors {
	return nil
}

type SyncNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
