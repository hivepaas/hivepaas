package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

// PasswordForgot this api handles request of resetting password from user by
// sending a reset link to the user's email address.
func (uc *UC) PasswordForgot(
	ctx context.Context,
	req *userdto.PasswordForgotReq,
) (*userdto.PasswordForgotResp, error) {
	user, err := uc.userRepo.GetByEmail(ctx, uc.db, req.Email,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil || user.IsDemoUser() {
		return nil, apperrors.Wrap(apperrors.ErrActionFailed)
	}

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, apperrors.New(apperrors.ErrActionNotAllowedByAdmin).
			WithMsgLog("user authentication method is enforce-sso")
	}

	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrActionNotAllowedByAdmin)
	}
	email, err := emailSetting.AsEmail()
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrActionFailed)
	}

	token, err := uc.userService.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrActionFailed)
	}

	resetLink := config.Current.DashboardPasswordResetURL(user.ID, token)
	err = uc.emailService.SendMailPasswordReset(ctx, uc.db, &emailservice.EmailDataPasswordReset{
		BaseTemplateData: emailservice.BaseTemplateData{
			Email:      email,
			Recipients: []string{user.Email},
			Subject:    "[LocalPaaS] Password reset",
		},
		ResetPasswordLink: resetLink,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrActionFailed)
	}

	return &userdto.PasswordForgotResp{}, nil
}
