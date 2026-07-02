package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdatePassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdatePasswordReq,
) (*userdto.UpdatePasswordResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.New(apperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
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

		err = uc.userService.ChangePassword(user, req.NewPassword, req.CurrentPassword)
		if err != nil {
			return apperrors.New(err).WithMsgLog("failed to change password")
		}

		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "password"),
		)
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &userdto.UpdatePasswordResp{}, nil
}
