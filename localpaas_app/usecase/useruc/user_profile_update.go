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
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdateProfile(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdateProfileReq,
) (*userdto.UpdateProfileResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		profileData := &userProfileData{}
		err := uc.loadUserProfileData(ctx, db, auth, req, profileData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingUserProfileData{}
		uc.preparePersistingUserProfileData(req, profileData, persistingData)

		return uc.persistUserProfileData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.UpdateProfileResp{}, nil
}

type userProfileData struct {
	User *entity.User
}

type persistingUserProfileData struct {
	UpdatingUser *entity.User
}

func (uc *UC) loadUserProfileData(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.UpdateProfileReq,
	data *userProfileData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if user.Status != base.UserStatusActive {
		return apperrors.New(apperrors.ErrActionNotAllowed).
			WithMsgLog("user '%s' not active", user.Email)
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
		if user.Email != "" {
			// When email of user exists, we don't allow changing
			return apperrors.New(apperrors.ErrEmailChangeUnallowed)
		}
		conflictUser, err := uc.userRepo.GetByEmail(ctx, db, req.Email)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrEmailUnavailable).
				WithMsgLog("email '%s' already exists", req.Email)
		}
	}

	// Save user photo to local disk
	if req.Photo != nil && req.Photo.FileExt != "" && len(req.Photo.DataBytes) > 0 {
		err = uc.userService.SaveUserPhoto(ctx, user, req.Photo.DataBytes, req.Photo.FileExt)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) preparePersistingUserProfileData(
	req *userdto.UpdateProfileReq,
	profileData *userProfileData,
	persistingData *persistingUserProfileData,
) {
	timeNow := timeutil.NowUTC()
	user := profileData.User
	persistingData.UpdatingUser = user

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
	if req.Photo != nil && req.Photo.Delete {
		user.Photo = ""
	}
	if req.Notes != nil {
		user.Notes = *req.Notes
	}
}

func (uc *UC) persistUserProfileData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingUserProfileData,
) error {
	err := uc.userRepo.Update(ctx, db, persistingData.UpdatingUser,
		bunex.UpdateColumns("updated_at", "username", "email", "full_name", "position", "photo", "notes"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
