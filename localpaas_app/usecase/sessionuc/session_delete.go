package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteSession(
	ctx context.Context,
	req *sessiondto.DeleteSessionReq,
) (resp *sessiondto.DeleteSessionResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.Del(ctx, req.User.AuthClaims.UserID, req.User.AuthClaims.UID)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.DeleteSessionResp{}, nil
}
