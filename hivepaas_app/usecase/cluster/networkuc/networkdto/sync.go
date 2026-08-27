package networkdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type SyncNetworkReq struct {
}

func NewSyncNetworkReq() *SyncNetworkReq {
	return &SyncNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SyncNetworkReq) Validate() hperrors.ValidationErrors {
	return nil
}

type SyncNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
