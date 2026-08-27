package sessionuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

func (uc *UC) GetCurrentUserByJWT(ctx context.Context, jwt string) (*basedto.User, error) {
	authClaims := &jwtsession.AuthClaims{}
	err := jwtsession.ParseToken(jwt, authClaims)
	if err != nil {
		if errors.Is(err, jwtsession.ErrTokenExpired) {
			return nil, hperrors.Wrap(hperrors.ErrSessionJWTExpired).WithCause(err)
		}
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	// Make sure the token is marked `existing` in redis
	if err = uc.userTokenRepo.Exist(ctx, authClaims.UserID, authClaims.UID); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, authClaims.UserID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &basedto.User{User: user, AuthClaims: authClaims}, nil
}

func (uc *UC) GetCurrentUserByAPIKey(ctx context.Context, keyID, secret string) (*basedto.User, error) {
	apiKeySetting, err := uc.settingRepo.GetByKind(ctx, uc.db, nil, base.SettingTypeAPIKey, keyID, false)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}
	if apiKeySetting == nil || !apiKeySetting.IsActive() {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}

	apiKey := apiKeySetting.MustAsAPIKey()
	if apiKey == nil {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}
	if err = apiKey.SecretKey.VerifyHash(secret); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}
	actingUserID := apiKeySetting.ObjectID

	user, err := uc.userService.LoadUser(ctx, uc.db, actingUserID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &basedto.User{User: user, AuthClaims: &jwtsession.AuthClaims{
		UserID:       user.ID,
		IsAPIKey:     true,
		AccessAction: apiKey.AccessAction,
	}}, nil
}
