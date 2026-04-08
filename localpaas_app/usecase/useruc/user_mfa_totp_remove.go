package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/totp"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) RemoveMFATotp(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.RemoveMFATotpReq,
) (*userdto.RemoveMFATotpResp, error) {
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
			return apperrors.New(apperrors.ErrActionNotAllowed).
				WithMsgLog("user authentication method is enforce-sso")
		}
		if user.SecurityOption == base.UserSecurityPassword2FA {
			return apperrors.New(apperrors.ErrActionNotAllowed).
				WithMsgLog("2FA is required by admin")
		}

		// Verify passcode
		if !totp.VerifyPasscode(req.Passcode, user.TotpSecret) {
			return apperrors.New(apperrors.ErrPasscodeMismatched)
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
