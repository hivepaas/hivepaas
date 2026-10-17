package sessionuc

import (
	"context"
	"errors"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	externalAvatarURLMaxLen = 2000
)

func (uc *UC) CreateOAuthSession(
	ctx context.Context,
	req *sessiondto.CreateOAuthSessionReq,
) (*sessiondto.CreateOAuthSessionResp, error) {
	oauthUser := req.User
	setting, oauth, err := uc.loadSSOSetting(ctx, oauthUser.Provider)
	if err != nil {
		return nil, err
	}
	email, err := uc.verifiedOAuthEmail(ctx, base.OAuthKind(setting.Kind), oauth, oauthUser)
	if err != nil {
		if errors.Is(err, hperrors.ErrOAuthEmailNotVerified) {
			if e := uc.recordUnverifiedOAuthEmail(ctx, oauthUser.Email); e != nil {
				return nil, e
			}
		}
		return nil, err
	}

	dbUser, err := uc.userRepo.GetByEmail(ctx, uc.db, email)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			if e := uc.recordLoginDenied(ctx, nil, auditMethodOAuth, auditReasonUnknownUser); e != nil {
				return nil, e
			}
			return nil, hperrors.Wrap(hperrors.ErrOAuthUserNotFound).WithCause(err)
		}
		return nil, hperrors.Wrap(err)
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
		return nil, hperrors.Wrap(err)
	}

	sessionResp, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{
		User:   dbUser,
		Method: auditMethodOAuth,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sessiondto.CreateOAuthSessionResp{
		BaseCreateSessionResp: *sessionResp,
	}, nil
}

// recordUnverifiedOAuthEmail records a sign-in refused for an email the provider does not vouch for against the user
// who has it, if one does: somebody may have typed that user's address into a profile to sign in as them.
func (uc *UC) recordUnverifiedOAuthEmail(ctx context.Context, email string) error {
	var user *entity.User
	if email = strutil.NormalizeEmail(email); email != "" {
		found, err := uc.userRepo.GetByEmail(ctx, uc.db, email)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		user = found
	}
	return uc.recordLoginDenied(ctx, user, auditMethodOAuth, auditReasonEmailNotVerified)
}
