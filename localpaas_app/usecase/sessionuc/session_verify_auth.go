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

	// Requested action is higher than the one limited within the session settings
	limitAccess := auth.User.AuthClaims.AccessAction
	if limitAccess != nil && !base.ActionAllowed(accessCheck.Action, *limitAccess) {
		return apperrors.New(apperrors.ErrUnauthorized).
			WithMsgLog("requested action is not allowed by session settings")
	}
	if auth.User.Role == base.UserRoleAdmin {
		return nil
	}

	if accessCheck.SubjectID == "" {
		accessCheck.SubjectType = base.SubjectTypeUser
		accessCheck.SubjectID = auth.User.ID
	}
	checkResult, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, accessCheck)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !checkResult {
		return apperrors.New(apperrors.ErrUnauthorized)
	}
	return nil
}
