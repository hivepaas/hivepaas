package useruc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdateUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdateUserReq,
) (*userdto.UpdateUserResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		userData := &userUpdateData{}
		err := uc.loadUserDataForUpdate(ctx, db, auth, req, userData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &userservice.PersistingUserData{}
		uc.prepareUpdatingUserData(req, userData, persistingData)

		return uc.userService.PersistUserData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.UpdateUserResp{}, nil
}

type userUpdateData struct {
	User *entity.User
}

func (uc *UC) loadUserDataForUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.UpdateUserReq,
	data *userUpdateData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.User = user

	// If username changes, need to verify the uniqueness
	if req.Username != "" && req.Username != user.Username {
		conflictUser, err := uc.userRepo.GetByUsername(ctx, db, req.Username)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrNameUnavailable).
				WithMsgLog("user '%s' already exists", req.Username)
		}
	}

	// If email changes, need to verify the uniqueness
	if req.Email != "" && req.Email != user.Email {
		conflictUser, err := uc.userRepo.GetByEmail(ctx, db, req.Email)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrEmailUnavailable).
				WithMsgLog("email '%s' already exists", req.Email)
		}
	}

	if req.Role != nil {
		if base.RoleCmp(auth.User.Role, *req.Role) < 0 {
			return apperrors.New(apperrors.ErrForbidden).
				WithMsgLog("you are not allowed to set a role higher than yours")
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingUserData(
	req *userdto.UpdateUserReq,
	updateData *userUpdateData,
	persistingData *userservice.PersistingUserData,
) {
	timeNow := timeutil.NowUTC()
	user := updateData.User

	user.UpdatedAt = timeNow
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Position != nil {
		user.Position = *req.Position
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Notes != nil {
		user.Notes = *req.Notes
	}
	if req.SecurityOption != nil {
		user.SecurityOption = *req.SecurityOption
	}
	if req.AccessExpireAt != nil {
		user.AccessExpireAt = *req.AccessExpireAt
	}

	if user.Status == base.UserStatusActive &&
		user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
		user.Status = base.UserStatusPending // User needs to set up 2FA authentication
	}

	persistingData.UpsertingUsers = append(persistingData.UpsertingUsers, user)

	if req.ModuleAccesses != nil {
		persistingData.DeletingAccesses = append(persistingData.DeletingAccesses,
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeModule,
			},
		)
		uc.preparePersistingUserModuleAccesses(user, req.ModuleAccesses, timeNow, persistingData)
	}
	if req.ProjectAccesses != nil {
		persistingData.DeletingAccesses = append(persistingData.DeletingAccesses,
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeProject,
			},
		)
		uc.preparePersistingUserProjectAccesses(user, req.ProjectAccesses, timeNow, persistingData)
	}
}
