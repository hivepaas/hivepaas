package sessionuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

func (uc *UC) GetCurrentUserByJWT(ctx context.Context, jwt string) (*basedto.User, error) {
	authClaims := &jwtsession.AuthClaims{}
	err := jwtsession.ParseToken(jwt, authClaims)
	if err != nil {
		if errors.Is(err, jwtsession.ErrTokenExpired) {
			return nil, apperrors.New(apperrors.ErrSessionJWTExpired).WithCause(err)
		}
		return nil, apperrors.New(apperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	// Make sure the token is marked `existing` in redis
	if err = uc.userTokenRepo.Exist(ctx, authClaims.UserID, authClaims.UID); err != nil {
		return nil, apperrors.New(apperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, authClaims.UserID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &basedto.User{User: user, AuthClaims: authClaims}, nil
}

func (uc *UC) GetCurrentUserByAPIKey(ctx context.Context, keyID, secret string) (*basedto.User, error) {
	apiKeySetting, err := uc.settingRepo.GetByKind(ctx, uc.db, nil, base.SettingTypeAPIKey, keyID, false)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrAPIKeyInvalid)
	}
	if apiKeySetting == nil || !apiKeySetting.IsActive() {
		return nil, apperrors.Wrap(apperrors.ErrAPIKeyInvalid)
	}

	apiKey := apiKeySetting.MustAsAPIKey()
	if apiKey == nil {
		return nil, apperrors.Wrap(apperrors.ErrAPIKeyMismatched)
	}
	if err = apiKey.SecretKey.VerifyHash(secret); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrAPIKeyMismatched)
	}
	actingUserID := apiKeySetting.ObjectID

	user, err := uc.userService.LoadUser(ctx, uc.db, actingUserID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &basedto.User{User: user, AuthClaims: &jwtsession.AuthClaims{
		UserID:       user.ID,
		IsAPIKey:     true,
		AccessAction: apiKey.AccessAction,
	}}, nil
}
