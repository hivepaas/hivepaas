package useruc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/totp"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) CompleteUserSignup(
	ctx context.Context,
	req *userdto.CompleteUserSignupReq,
) (*userdto.CompleteUserSignupResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		signupData := &userSignupData{}
		err := uc.loadUserSignupData(ctx, db, req, signupData)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingUserSignupData{}
		if err = uc.preparePersistingUserSignupData(req, signupData, persistingData); err != nil {
			return err
		}

		return uc.persistUserSignupData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &userdto.CompleteUserSignupResp{}, nil
}

type userSignupData struct {
	User *entity.User
}

type persistingUserSignupData struct {
	persistingUserProfileData
}

func (uc *UC) loadUserSignupData(
	ctx context.Context,
	db database.IDB,
	req *userdto.CompleteUserSignupReq,
	data *userSignupData,
) error {
	inviteToken, err := uc.userService.ParseUserInviteToken(req.InviteToken)
	if err != nil {
		return apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}

	user, err := uc.userRepo.GetByID(ctx, db, inviteToken.UserID,
		bunex.SelectFor("UPDATE OF \"user\""),
		bunex.SelectRelationIf(req.Photo.IsChanged(), "PhotoData"),
	)
	if err != nil {
		return apperrors.New(err)
	}

	if user.Status != base.UserStatusPending {
		return apperrors.New(apperrors.ErrActionNotAllowed).
			WithMsgLog("user '%s' not require signup", user.Email)
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

	return nil
}

func (uc *UC) preparePersistingUserSignupData(
	req *userdto.CompleteUserSignupReq,
	signupData *userSignupData,
	persistingData *persistingUserSignupData,
) error {
	timeNow := timeutil.NowUTC()
	user := signupData.User

	user.UpdatedAt = timeNow
	user.Username = req.Username
	user.FullName = req.FullName
	user.Status = base.UserStatusActive

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		user.Password = ""
	} else {
		if req.Password == "" {
			return apperrors.NewArgumentInvalid("Password").
				WithMsgLog("password is required")
		}
		err := uc.userService.ChangePassword(user, req.Password, userservice.SkipCheckingCurrentPassword)
		if err != nil {
			return apperrors.New(err)
		}
	}
	if user.SecurityOption == base.UserSecurityPassword2FA {
		if req.Passcode == "" || req.MFATotpSecret == "" {
			return apperrors.NewArgumentInvalid("Passcode").
				WithMsgLog("passcode and totp secret are required")
		}
		if !totp.VerifyPasscode(req.Passcode, req.MFATotpSecret) {
			return apperrors.New(apperrors.ErrPasscodeMismatched)
		}
		user.TotpSecret = req.MFATotpSecret
	}

	if req.Photo.IsChanged() {
		uc.preparePersistingUserPhoto(req.Photo, user, timeNow, &persistingData.persistingUserProfileData)
	}

	persistingData.UpdatingUser = user
	return nil
}

func (uc *UC) persistUserSignupData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingUserSignupData,
) error {
	return uc.persistUserProfileData(ctx, db, &persistingData.persistingUserProfileData)
}
