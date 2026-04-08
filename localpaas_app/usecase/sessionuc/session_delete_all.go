package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteAllSessions(
	ctx context.Context,
	req *sessiondto.DeleteAllSessionsReq,
) (resp *sessiondto.DeleteAllSessionsResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.DelAll(ctx, req.User.AuthClaims.UserID)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.DeleteAllSessionsResp{}, nil
}
