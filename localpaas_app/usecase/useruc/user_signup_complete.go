package useruc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/totp"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) CompleteUserSignup(
	ctx context.Context,
	req *userdto.CompleteUserSignupReq,
) (*userdto.CompleteUserSignupResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		signupData := &userSignupData{}
		err := uc.loadUserSignupData(ctx, db, req, signupData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingUserSignupData{}
		if err = uc.preparePersistingUserSignupData(req, signupData, persistingData); err != nil {
			return err
		}

		return uc.persistUserSignupData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.CompleteUserSignupResp{}, nil
}

type userSignupData struct {
	User *entity.User
}

type persistingUserSignupData struct {
	UpdatingUser *entity.User
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
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}
		if conflictUser != nil {
			return apperrors.New(apperrors.ErrNameUnavailable).
				WithMsgLog("user '%s' already exists", req.Username)
		}
	}

	// Save user photo to local disk
	err = uc.userService.SaveUserPhoto(ctx, user, req.Photo.DataBytes, req.Photo.FileExt)
	if err != nil {
		return apperrors.Wrap(err)
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
	persistingData.UpdatingUser = user

	user.UpdatedAt = timeNow
	user.Username = req.Username
	user.FullName = req.FullName
	user.Status = base.UserStatusActive

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		user.Password = ""
	} else {
		if req.Password == "" {
			return apperrors.NewParamInvalid("Password").
				WithMsgLog("password is required")
		}
		if err := uc.userService.ChangePassword(user, req.Password, userservice.SkipCheckingCurrentPassword); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if user.SecurityOption == base.UserSecurityPassword2FA {
		if req.Passcode == "" || req.MFATotpSecret == "" {
			return apperrors.NewParamInvalid("Passcode").
				WithMsgLog("passcode and totp secret are required")
		}
		if !totp.VerifyPasscode(req.Passcode, req.MFATotpSecret) {
			return apperrors.New(apperrors.ErrPasscodeMismatched)
		}
		user.TotpSecret = req.MFATotpSecret
	}

	return nil
}

func (uc *UC) persistUserSignupData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingUserSignupData,
) error {
	err := uc.userRepo.Update(ctx, db, persistingData.UpdatingUser)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
