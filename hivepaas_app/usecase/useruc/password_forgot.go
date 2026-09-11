package useruc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
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
		// Recorded when it is a refusal rather than the install failing, and
		// without the address that was typed - see recordUserChangeDenied. An
		// address nobody has, tried repeatedly, is somebody looking for accounts.
		if reason := forgotRefusalReason(user, err); reason != "" {
			if e := uc.recordUserChangeDenied(ctx, uc.db, nil,
				auditSectionPasswordResetRequest, user, reason); e != nil {
				return nil, hperrors.Wrap(e)
			}
		}
		return nil, hperrors.Wrap(hperrors.ErrActionFailed)
	}

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		if e := uc.recordUserChangeDenied(ctx, uc.db, nil,
			auditSectionPasswordResetRequest, user, auditReasonNotAllowed); e != nil {
			return nil, hperrors.Wrap(e)
		}
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowedByAdmin).
			WithMsgLog("user authentication method is enforce-sso")
	}

	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowedByAdmin)
	}
	email, err := emailSetting.AsEmail()
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrActionFailed)
	}

	token, err := uc.userService.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrActionFailed)
	}

	resetLink := config.Current().DashboardPasswordResetURL(user.ID, token)
	err = uc.emailService.SendMailPasswordReset(ctx, uc.db, &emailservice.EmailDataPasswordReset{
		BaseTemplateData: emailservice.BaseTemplateData{
			Email:      email,
			Recipients: []string{user.Email},
			Subject:    "[HivePaaS] Password reset",
		},
		ResetPasswordLink: resetLink,
	})
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrActionFailed)
	}

	// After the mail, because what is recorded is that a link went out - and with
	// no actor, because the form asks for an address and nothing else.
	if err = uc.recordPasswordResetRequest(ctx, uc.db, user); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.PasswordForgotResp{}, nil
}

// forgotRefusalReason names why the form was turned down, and answers empty for
// anything that is not a refusal - the database being unreachable is the install
// failing, not somebody being turned away.
func forgotRefusalReason(user *entity.User, err error) string {
	switch {
	case errors.Is(err, hperrors.ErrNotFound):
		return auditReasonUnknownEmail
	case err == nil && user != nil:
		// Got as far as a real account, so what refused it was the demo check.
		return auditReasonNotAllowed
	}
	return ""
}
