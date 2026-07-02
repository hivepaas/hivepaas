package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteSessionReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteSessionReq() *DeleteSessionReq {
	return &DeleteSessionReq{}
}

func (req *DeleteSessionReq) Validate() apperrors.ValidationErrors {
	return nil
}

type DeleteSessionResp struct {
	Meta *basedto.Meta `json:"meta"`
}
