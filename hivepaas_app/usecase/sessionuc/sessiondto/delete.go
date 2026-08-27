package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteSessionReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteSessionReq() *DeleteSessionReq {
	return &DeleteSessionReq{}
}

func (req *DeleteSessionReq) Validate() hperrors.ValidationErrors {
	return nil
}

type DeleteSessionResp struct {
	Meta *basedto.Meta `json:"meta"`
}
