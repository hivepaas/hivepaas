package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) ResetPassword(
	ctx context.Context,
	req *userdto.ResetPasswordReq,
) (*userdto.ResetPasswordResp, error) {
	tokenClaims, err := uc.userService.ParsePasswordResetToken(req.Token)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, tokenClaims.UserID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if user.IsDemoUser() {
			return hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
		}

		err = uc.userService.ChangePassword(user, req.Password, userservice.SkipCheckingCurrentPassword)
		if err != nil {
			return hperrors.Wrap(err).WithMsgLog("failed to change password")
		}

		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "password"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.ResetPasswordResp{}, nil
}
