package useruc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
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
