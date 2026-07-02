package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

// RequestResetPassword this api handles request of resetting password from admins by
// returning a reset link to the admin or sending it to the target user's email address.
func (uc *UC) RequestResetPassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.RequestResetPasswordReq,
) (*userdto.RequestResetPasswordResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.New(apperrors.ErrUserDemoUnauthorized)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
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
			return nil, apperrors.New(err)
		}

		email, err := emailSetting.AsEmail()
		if err != nil {
			return nil, apperrors.New(err)
		}

		err = uc.emailService.SendMailPasswordReset(ctx, uc.db, &emailservice.EmailDataPasswordReset{
			BaseTemplateData: emailservice.BaseTemplateData{
				Email:      email,
				Recipients: []string{user.Email},
				Subject:    "[HivePaaS] Password reset",
			},
			ResetPasswordLink: resetLink,
		})
		if err != nil {
			return nil, apperrors.New(err)
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
