package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) RequestResetPassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.RequestResetPasswordReq,
) (*userdto.RequestResetPasswordResp, error) {
	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, apperrors.New(apperrors.ErrActionNotAllowed).
			WithMsgLog("user authentication method is enforce-sso")
	}

	token, err := uc.userService.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to generate password reset token")
	}

	resetLink := config.Current.DashboardPasswordResetURL(user.ID, token)

	if req.SendResettingEmail {
		emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		email, err := emailSetting.AsEmail()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		err = uc.emailService.SendMailPasswordReset(ctx, uc.db, &emailservice.EmailDataPasswordReset{
			Email:             email,
			Recipients:        []string{user.Email},
			ResetPasswordLink: resetLink,
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		// When send the link via email, we don't return it via the response
		resetLink = ""
	}

	return &userdto.RequestResetPasswordResp{
		Data: &userdto.RequestResetPasswordDataResp{
			ResetPasswordLink: resetLink,
		},
	}, nil
}
