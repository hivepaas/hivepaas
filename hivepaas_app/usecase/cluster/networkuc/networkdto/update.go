package networkdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateNetworkReq struct {
	settings.UpdateSettingReq
}

func NewUpdateNetworkReq() *UpdateNetworkReq {
	return &UpdateNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNetworkReq) Validate() apperrors.ValidationErrors {
	return nil
}

type UpdateNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
