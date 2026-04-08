package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) ResetPassword(
	ctx context.Context,
	req *userdto.ResetPasswordReq,
) (*userdto.ResetPasswordResp, error) {
	tokenClaims, err := uc.userService.ParsePasswordResetToken(req.Token)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, tokenClaims.UserID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.userService.ChangePassword(user, req.Password, userservice.SkipCheckingCurrentPassword)
		if err != nil {
			return apperrors.New(err).WithMsgLog("failed to change password")
		}

		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "password"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.ResetPasswordResp{}, nil
}
