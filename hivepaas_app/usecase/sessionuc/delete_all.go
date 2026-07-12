package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteAllSessions(
	ctx context.Context,
	req *sessiondto.DeleteAllSessionsReq,
) (resp *sessiondto.DeleteAllSessionsResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.DelAll(ctx, req.User.AuthClaims.UserID)
	if err != nil {
		return nil, apperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.DeleteAllSessionsResp{}, nil
}
