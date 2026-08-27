package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (uc *UC) VerifyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	accessCheck permission.AccessCheck,
) error {
	if auth.User.AuthClaims.IsRefresh {
		return hperrors.Wrap(hperrors.ErrForbidden).
			WithMsgLog("refresh token is not allowed")
	}
	if accessCheck == nil {
		return nil
	}
	if !accessCheck.IsValid() {
		return hperrors.NewArgumentInvalid("Either 'Action' or 'AllOf' or 'AnyOf'")
	}

	// Requested action is higher than the one limited within the session settings
	limitAccess := auth.User.AuthClaims.AccessAction
	if limitAccess != nil {
		allowed := false
		baseCheck := accessCheck.GetBase()
		switch {
		case baseCheck.Action != "":
			allowed = limitAccess.Allows(baseCheck.Action)
		case len(baseCheck.AllOf) > 0:
			allowed = limitAccess.AllowsAll(baseCheck.AllOf)
		case len(baseCheck.AnyOf) > 0:
			allowed = limitAccess.AllowsAny(baseCheck.AnyOf)
		}
		if !allowed {
			if auth.User.IsDemoUser() { // Special case: demo user
				return hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
			}
			return hperrors.Wrap(hperrors.ErrUnauthorized).
				WithMsgLog("requested action is not allowed by session settings")
		}
	}

	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, accessCheck)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !hasPerm {
		return hperrors.Wrap(hperrors.ErrUnauthorized)
	}
	return nil
}
