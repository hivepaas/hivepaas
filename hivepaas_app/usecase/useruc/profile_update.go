package useruc

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdateProfile(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdateProfileReq,
) (*userdto.UpdateProfileResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		profileData := &userProfileData{}
		err := uc.loadUserProfileData(ctx, db, auth, req, profileData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingUserProfileData{}
		uc.preparePersistingUserProfileData(req, profileData, persistingData)

		return uc.persistUserProfileData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.UpdateProfileResp{}, nil
}

type userProfileData struct {
	User *entity.User
}

type persistingUserProfileData struct {
	UpdatingUser             *entity.User
	UpsertingBinObjects      []*entity.BinObject
	HardDeletingBinObjectIDs []string
}

func (uc *UC) loadUserProfileData(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.UpdateProfileReq,
	data *userProfileData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
		bunex.SelectFor("UPDATE OF \"user\""),
		bunex.SelectRelationIf(req.Photo.IsChanged(), "PhotoData"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if user.Status != base.UserStatusActive {
		return hperrors.Wrap(hperrors.ErrActionNotAllowed).
			WithMsgLog("user '%s' not active", user.Email)
	}
	data.User = user

	// If username changes, need to verify the uniqueness
	if req.Username != "" && req.Username != user.Username {
		conflictUser, err := uc.userRepo.GetByUsername(ctx, db, req.Username)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if conflictUser != nil {
			return hperrors.Wrap(hperrors.ErrUsernameUnavailable).
				WithMsgLog("user '%s' already exists", req.Username)
		}
	}

	// If email changes, need to verify the uniqueness
	if req.Email != "" && req.Email != user.Email {
		if user.Email != "" {
			// When email of user exists, we don't allow changing
			return hperrors.Wrap(hperrors.ErrEmailChangeUnallowed)
		}
		conflictUser, err := uc.userRepo.GetByEmail(ctx, db, req.Email)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if conflictUser != nil {
			return hperrors.Wrap(hperrors.ErrEmailUnavailable).
				WithMsgLog("email '%s' already exists", req.Email)
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
	if req.Notes != nil {
		user.Notes = *req.Notes
	}
	if req.Photo.IsChanged() {
		uc.preparePersistingUserPhoto(req.Photo, user, timeNow, persistingData)
	}

	persistingData.UpdatingUser = user
}

func (uc *UC) preparePersistingUserPhoto(
	req *userdto.UserPhotoReq,
	user *entity.User,
	timeNow time.Time,
	persistingData *persistingUserProfileData,
) {
	if !req.IsChanged() {
		return
	}
	photoData := user.PhotoData

	if photoData != nil && photoData.ID != "" {
		// User photo may take a remarkable space, so we hard-delete it
		persistingData.HardDeletingBinObjectIDs = append(persistingData.HardDeletingBinObjectIDs, photoData.ID)
	}

	if req.Delete {
		user.Photo = ""
		return
	}

	photoData = &entity.BinObject{
		ID:          gofn.Must(ulid.NewStringULID()),
		Type:        base.BinObjectTypeObjectIcon,
		Status:      base.BinObjectStatusActive,
		Name:        req.FileName,
		ContentType: fileutil.TypeByExtension(filepath.Ext(req.FileName)),
		Data:        req.DataBytes,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	user.Photo = photoData.ID
	persistingData.UpsertingBinObjects = append(persistingData.UpsertingBinObjects, photoData)
}

func (uc *UC) persistUserProfileData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingUserProfileData,
) error {
	err := uc.userRepo.Update(ctx, db, persistingData.UpdatingUser)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.binObjectRepo.DeleteByIDs(ctx, db, persistingData.HardDeletingBinObjectIDs,
		bunex.DeleteWithForceDelete())
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
