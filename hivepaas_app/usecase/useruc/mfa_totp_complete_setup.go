package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/totp"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) CompleteMFATotpSetup(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.CompleteMFATotpSetupReq,
) (*userdto.CompleteMFATotpSetupResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.New(apperrors.ErrUserDemoUnauthorized)
	}

	mfaTokenClaims, err := uc.userService.ParseMFATotpSetupToken(req.TotpToken)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.New(err)
		}
		if user.SecurityOption == base.UserSecurityEnforceSSO {
			return apperrors.New(apperrors.ErrActionNotAllowed).
				WithMsgLog("user authentication method is enforce-sso")
		}

		// Verify passcode
		if !totp.VerifyPasscode(req.Passcode, mfaTokenClaims.Secret) {
			return apperrors.New(apperrors.ErrPasscodeMismatched)
		}

		user.TotpSecret = mfaTokenClaims.Secret
		if user.Status == base.UserStatusPending && user.SecurityOption == base.UserSecurityPassword2FA {
			user.Status = base.UserStatusActive
		}
		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "totp_secret", "status"),
		)
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &userdto.CompleteMFATotpSetupResp{}, nil
}
