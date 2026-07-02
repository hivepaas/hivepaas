package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetLoginOptionsReq struct {
}

func NewGetLoginOptionsReq() *GetLoginOptionsReq {
	return &GetLoginOptionsReq{}
}

func (req *GetLoginOptionsReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetLoginOptionsResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data []*LoginOptionResp `json:"data"`
}

type LoginOptionResp struct {
	Type    base.OAuthKind `json:"type"`
	Name    string         `json:"name"`
	Icon    string         `json:"icon"`
	AuthURL string         `json:"authURL"`
}
