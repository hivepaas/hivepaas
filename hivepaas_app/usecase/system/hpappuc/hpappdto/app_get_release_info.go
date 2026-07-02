package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

type GetHpAppReleaseInfoReq struct {
}

func NewGetHpAppReleaseInfoReq() *GetHpAppReleaseInfoReq {
	return &GetHpAppReleaseInfoReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetHpAppReleaseInfoReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetHpAppReleaseInfoResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *HpAppReleaseInfoResp `json:"data"`
}

type HpAppReleaseInfoResp struct {
	*hpappservice.AppReleaseInfo
}
