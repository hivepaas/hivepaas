package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteAllSessions(
	ctx context.Context,
	req *sessiondto.DeleteAllSessionsReq,
) (resp *sessiondto.DeleteAllSessionsResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.DelAll(ctx, req.User.AuthClaims.UserID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	// After the tokens are dead, so what is recorded is what happened. Worth more
	// than the single logout next door: this is the act that ends every other
	// session the account has, wherever they are.
	if err = uc.recordLogout(ctx, req.User, auditSectionLogoutAll); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sessiondto.DeleteAllSessionsResp{}, nil
}
