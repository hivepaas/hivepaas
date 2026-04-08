package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) LoginPasswordForgot(
	ctx context.Context,
	req *sessiondto.LoginPasswordForgotReq,
) (*sessiondto.LoginPasswordForgotResp, error) {
	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil {
		return nil, apperrors.NewNotFound("System email setting")
	}

	user, err := uc.userRepo.GetByUsernameOrEmail(ctx, uc.db, req.Email, req.Email,
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

	return &sessiondto.LoginPasswordForgotResp{}, nil
}
