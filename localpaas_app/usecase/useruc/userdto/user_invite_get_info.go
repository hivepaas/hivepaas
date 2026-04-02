package userdto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetUserInviteInfoReq struct {
}

func NewGetUserInviteInfoReq() *GetUserInviteInfoReq {
	return &GetUserInviteInfoReq{}
}

func (req *GetUserInviteInfoReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetUserInviteInfoResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *UserInviteInfoResp `json:"data"`
}

type UserInviteInfoResp struct {
	CanSendInviteEmails bool `json:"canSendInviteEmails"`
}
