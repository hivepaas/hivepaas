package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) DeleteUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.DeleteUserReq,
) (*userdto.DeleteUserResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		userData := &deleteUserData{}
		err := uc.loadUserDataForDelete(ctx, db, auth, req, userData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &userservice.PersistingUserData{}
		uc.prepareDeletingUser(userData, persistingData)

		err = uc.userService.DeleteUser(ctx, db, userData.User)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return uc.userService.PersistUserData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.DeleteUserResp{}, nil
}

type deleteUserData struct {
	User *entity.User
}

func (uc *UC) loadUserDataForDelete(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.DeleteUserReq,
	data *deleteUserData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.User = user

	if user.Role == base.UserRoleAdmin {
		if auth.User.Role != base.UserRoleAdmin {
			return apperrors.Wrap(apperrors.ErrActionNotAllowed).
				WithMsgLog("member user cannot delete admin user")
		}

		// Make sure there is at least one active admin user in the system after the deletion
		otherAdmins, _, err := uc.userRepo.List(ctx, db, nil,
			bunex.SelectWhere("id != ?", user.ID),
			bunex.SelectWhere("role = ?", base.UserRoleAdmin),
			bunex.SelectWhere("status = ?", base.UserStatusActive),
			bunex.SelectWhere("access_expire_at IS NULL OR access_expire_at > NOW()"),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(otherAdmins) == 0 {
			return apperrors.Wrap(apperrors.ErrActionNotAllowed).
				WithMsgLog("cannot delete the last admin user")
		}
	}

	return nil
}

func (uc *UC) prepareDeletingUser(
	userData *deleteUserData,
	persistingData *userservice.PersistingUserData,
) {
	user := userData.User
	user.DeletedAt = timeutil.NowUTC()

	persistingData.UpsertingUsers = append(persistingData.UpsertingUsers, user)
}
