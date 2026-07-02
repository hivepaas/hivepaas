package sessiondto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteAllSessionsReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteAllSessionsReq() *DeleteAllSessionsReq {
	return &DeleteAllSessionsReq{}
}

func (req *DeleteAllSessionsReq) Validate() apperrors.ValidationErrors {
	return nil
}

type DeleteAllSessionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
