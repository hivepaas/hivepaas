package volumedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type SyncVolumeReq struct {
}

func NewSyncVolumeReq() *SyncVolumeReq {
	return &SyncVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SyncVolumeReq) Validate() hperrors.ValidationErrors {
	return nil
}

type SyncVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
