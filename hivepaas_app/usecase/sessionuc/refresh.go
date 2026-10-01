package sessionuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) RefreshSession(
	ctx context.Context,
	user *basedto.User,
) (resp *sessiondto.RefreshSessionResp, err error) {
	// JWT token must be refresh token
	if !user.AuthClaims.IsRefresh {
		return nil, hperrors.Wrap(hperrors.ErrSessionRefreshTokenRequired)
	}

	// The start travels with the session, so its deadline is measured from the
	// login and not from this renewal. A session minted before that claim existed
	// carries nothing here and is dated from now - one more full window, once.
	createReq := &sessiondto.BaseCreateSessionReq{
		User:      user.User,
		StartedAt: user.AuthClaims.SessionStartedAt(),
		Method:    auditMethodRefresh,
	}
	if user.AuthClaims.IsAPIKey {
		var keySetting *entity.Setting
		if user.AuthClaims.APIKeyID != "" {
			keySetting, err = uc.settingRepo.GetByID(ctx, uc.db, nil, base.SettingTypeAPIKey,
				user.AuthClaims.APIKeyID, false)
			if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
				return nil, hperrors.Wrap(err)
			}
		}
		if err = renewAPIKeySession(createReq, user, keySetting); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	sessionData, err := uc.createSession(ctx, createReq)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create session")
	}

	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.Del(ctx, user.AuthClaims.UserID, user.AuthClaims.UID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.RefreshSessionResp{
		Data: &sessiondto.RefreshSessionDataResp{BaseCreateSessionResp: sessionData},
	}, nil
}

// renewAPIKeySession makes the renewal of a session of a key a session of the
// same key, as the key is now: its limits carried over - left out, the renewal
// was a session of the person, with all they may do - and a key revoked,
// expired, or another's ending the session instead. A session signed in before
// the key's id was carried has no key to read, and signs in again.
func renewAPIKeySession(req *sessiondto.BaseCreateSessionReq, user *basedto.User, keySetting *entity.Setting) error {
	if keySetting == nil || !keySetting.IsActive() || keySetting.ObjectID != user.ID {
		return hperrors.Wrap(hperrors.ErrAPIKeyInvalid).
			WithMsgLog("a session of a key is renewed only while the key is valid: sign in with it again")
	}
	apiKey := keySetting.MustAsAPIKey()
	req.IsAPIKey = true
	req.APIKeyID = keySetting.ID
	req.AccessAction = apiKey.AccessAction
	req.Capabilities = apiKey.Capabilities
	return nil
}
