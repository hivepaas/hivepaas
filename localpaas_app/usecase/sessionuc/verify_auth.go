package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/permission"
)

func (uc *UC) VerifyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	accessCheck *permission.AccessCheck,
) error {
	if auth.User.AuthClaims.IsRefresh {
		return apperrors.New(apperrors.ErrForbidden).
			WithMsgLog("refresh token is not allowed")
	}
	if accessCheck == nil {
		return nil
	}
	if !accessCheck.IsValid() {
		return apperrors.NewParamInvalid("Either 'Action' or 'AllOf' or 'AnyOf'")
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
				return apperrors.New(apperrors.ErrUserDemoUnauthorized)
			}
			return apperrors.New(apperrors.ErrUnauthorized).
				WithMsgLog("requested action is not allowed by session settings")
		}
	}

	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return nil
	}

	if accessCheck.SubjectID == "" {
		accessCheck.SubjectType = base.SubjectTypeUser
		accessCheck.SubjectID = auth.User.ID
	}
	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, accessCheck)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !hasPerm {
		return apperrors.New(apperrors.ErrUnauthorized)
	}
	return nil
}
