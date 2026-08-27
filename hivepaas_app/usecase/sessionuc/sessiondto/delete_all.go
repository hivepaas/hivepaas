package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteAllSessionsReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteAllSessionsReq() *DeleteAllSessionsReq {
	return &DeleteAllSessionsReq{}
}

func (req *DeleteAllSessionsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type DeleteAllSessionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
