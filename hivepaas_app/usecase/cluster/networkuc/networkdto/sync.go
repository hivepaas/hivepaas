package networkdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
