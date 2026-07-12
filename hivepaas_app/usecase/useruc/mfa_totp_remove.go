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

func (uc *UC) RemoveMFATotp(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.RemoveMFATotpReq,
) (*userdto.RemoveMFATotpResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.Wrap(apperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if user.TotpSecret == "" {
			return nil
		}
		if user.SecurityOption == base.UserSecurityEnforceSSO {
			return apperrors.Wrap(apperrors.ErrActionNotAllowed).
				WithMsgLog("user authentication method is enforce-sso")
		}
		if user.SecurityOption == base.UserSecurityPassword2FA {
			return apperrors.Wrap(apperrors.ErrActionNotAllowed).
				WithMsgLog("2FA is required by admin")
		}

		// Verify passcode
		if !totp.VerifyPasscode(req.Passcode, user.TotpSecret) {
			return apperrors.Wrap(apperrors.ErrPasscodeMismatched)
		}

		user.TotpSecret = ""
		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "totp_secret"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.RemoveMFATotpResp{}, nil
}
