package sessiondto

import "github.com/hivepaas/hivepaas/hivepaas_app/basedto"

type RefreshSessionResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *RefreshSessionDataResp `json:"data"`
}

type RefreshSessionDataResp struct {
	*BaseCreateSessionResp
}
