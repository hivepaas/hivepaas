package volumedto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type SyncVolumeReq struct {
}

func NewSyncVolumeReq() *SyncVolumeReq {
	return &SyncVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SyncVolumeReq) Validate() apperrors.ValidationErrors {
	return nil
}

type SyncVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
