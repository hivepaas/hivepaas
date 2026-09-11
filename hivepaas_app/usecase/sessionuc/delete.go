package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteSession(
	ctx context.Context,
	req *sessiondto.DeleteSessionReq,
) (resp *sessiondto.DeleteSessionResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.Del(ctx, req.User.AuthClaims.UserID, req.User.AuthClaims.UID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	// After the token is dead, so what is recorded is what happened.
	if err = uc.recordLogout(ctx, req.User, auditSectionLogoutCurrent); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sessiondto.DeleteSessionResp{}, nil
}
