package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (uc *UC) VerifyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	accessCheck *permission.AccessCheck,
) error {
	if auth.User.AuthClaims.IsRefresh {
		return apperrors.Wrap(apperrors.ErrForbidden).
			WithMsgLog("refresh token is not allowed")
	}
	if accessCheck == nil {
		return nil
	}
	if !accessCheck.IsValid() {
		return apperrors.NewArgumentInvalid("Either 'Action' or 'AllOf' or 'AnyOf'")
	}

	// Requested action is higher than the one limited within the session settings
	limitAccess := auth.User.AuthClaims.AccessAction
	if limitAccess != nil {
		allowed := false
		switch {
		case accessCheck.Action != "":
			allowed = limitAccess.Allows(accessCheck.Action)
		case len(accessCheck.AllOf) > 0:
			allowed = limitAccess.AllowsAll(accessCheck.AllOf)
		case len(accessCheck.AnyOf) > 0:
			allowed = limitAccess.AllowsAny(accessCheck.AnyOf)
		}
		if !allowed {
			if auth.User.IsDemoUser() { // Special case: demo user
				return apperrors.Wrap(apperrors.ErrUserDemoUnauthorized)
			}
			return apperrors.Wrap(apperrors.ErrUnauthorized).
				WithMsgLog("requested action is not allowed by session settings")
		}
	}

	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, accessCheck)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !hasPerm {
		return apperrors.Wrap(apperrors.ErrUnauthorized)
	}
	return nil
}
