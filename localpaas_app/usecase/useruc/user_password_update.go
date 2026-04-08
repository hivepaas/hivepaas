package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdatePassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdatePasswordReq,
) (*userdto.UpdatePasswordResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if user.SecurityOption == base.UserSecurityEnforceSSO {
			return apperrors.New(apperrors.ErrActionNotAllowed).
				WithMsgLog("user authentication method is enforce-sso")
		}

		err = uc.userService.ChangePassword(user, req.NewPassword, req.CurrentPassword)
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

	return &userdto.UpdatePasswordResp{}, nil
}
