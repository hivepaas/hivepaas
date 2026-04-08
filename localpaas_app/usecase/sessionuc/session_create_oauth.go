package sessionuc

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

const (
	externalAvatarURLMaxLen = 2000
)

func (uc *UC) CreateOAuthSession(
	ctx context.Context,
	req *sessiondto.CreateOAuthSessionReq,
) (*sessiondto.CreateOAuthSessionResp, error) {
	oauthUser := req.User
	email := oauthUser.Email
	if email == "" {
		return nil, apperrors.New(apperrors.ErrInternalServer).
			WithMsgLog("unable to create oauth session, email is not returned from provider")
	}

	dbUser, err := uc.userRepo.GetByEmail(ctx, uc.db, email)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Make synchronization as info of user may be changed in the IdP system
	updateCols := make(map[string]struct{})
	if strings.HasPrefix(dbUser.Photo, "http") && len(oauthUser.AvatarURL) < externalAvatarURLMaxLen {
		dbUser.Photo = oauthUser.AvatarURL
		dbUser.UpdatedAt = timeutil.NowUTC()
		updateCols["photo"] = struct{}{}
		updateCols["updated_at"] = struct{}{}
	}
	if dbUser.Status == base.UserStatusPending && dbUser.LastAccess.IsZero() {
		dbUser.Status = base.UserStatusActive
		dbUser.UpdatedAt = timeutil.NowUTC()
		dbUser.LastAccess = dbUser.UpdatedAt
		updateCols["status"] = struct{}{}
		updateCols["updated_at"] = struct{}{}
		updateCols["last_access"] = struct{}{}
	}
	// Saves the user
	err = uc.userRepo.Update(ctx, uc.db, dbUser, bunex.UpdateColumns(gofn.MapKeys(updateCols)...))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	sessionResp, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: dbUser})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sessiondto.CreateOAuthSessionResp{
		BaseCreateSessionResp: *sessionResp,
	}, nil
}
