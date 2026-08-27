package userdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetUserInviteInfoReq struct {
}

func NewGetUserInviteInfoReq() *GetUserInviteInfoReq {
	return &GetUserInviteInfoReq{}
}

func (req *GetUserInviteInfoReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetUserInviteInfoResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *UserInviteInfoResp `json:"data"`
}

type UserInviteInfoResp struct {
	CanSendInviteEmails bool `json:"canSendInviteEmails"`
}
