package useruc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) GetUserInviteInfo(
	ctx context.Context,
	_ *basedto.Auth,
	_ *userdto.GetUserInviteInfoReq,
) (*userdto.GetUserInviteInfoResp, error) {
	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.GetUserInviteInfoResp{
		Data: &userdto.UserInviteInfoResp{
			CanSendInviteEmails: emailSetting != nil && emailSetting.IsActive(),
		},
	}, nil
}
