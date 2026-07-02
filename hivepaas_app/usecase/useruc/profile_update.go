package useruc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		return nil, apperrors.New(apperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		profileData := &userProfileData{}
		err := uc.loadUserProfileData(ctx, db, auth, req, profileData)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingUserProfileData{}
		uc.preparePersistingUserProfileData(req, profileData, persistingData)

		return uc.persistUserProfileData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
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
		return apperrors.New(err)
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
			return apperrors.New(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrUsernameUnavailable).
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
			return apperrors.New(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrEmailUnavailable).
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

	if req.Delete {
		if photoData != nil && photoData.ID != "" {
			// User photo may take a remarkable space, so we hard-delete it
			persistingData.HardDeletingBinObjectIDs = append(persistingData.HardDeletingBinObjectIDs, photoData.ID)
		}
		user.Photo = ""
		user.PhotoID = ""
		return
	}

	if photoData == nil {
		photoData = &entity.BinObject{
			ID:        gofn.Must(ulid.NewStringULID()),
			CreatedAt: timeNow,
		}
	}
	fileExt := strings.ToLower(filepath.Ext(req.FileName))
	photoData.UpdatedAt = timeNow
	photoData.Type = base.BinObjectTypeUserPhoto
	photoData.Status = base.BinObjectStatusActive
	photoData.Name = req.FileName
	photoData.ContentType = fileutil.TypeByExtension(fileExt)
	photoData.Data = req.DataBytes
	persistingData.UpsertingBinObjects = append(persistingData.UpsertingBinObjects, photoData)

	user.PhotoID = photoData.ID
	user.Photo = fmt.Sprintf("%v/images/%v-%v", config.Current.HTTPServer.BasePath,
		user.PhotoID, rand.Int31n(1000)) //nolint
}

func (uc *UC) persistUserProfileData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingUserProfileData,
) error {
	err := uc.userRepo.Update(ctx, db, persistingData.UpdatingUser)
	if err != nil {
		return apperrors.New(err)
	}

	err = uc.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	err = uc.binObjectRepo.DeleteByIDs(ctx, db, persistingData.HardDeletingBinObjectIDs,
		bunex.DeleteWithForceDelete())
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
