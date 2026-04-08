package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) GetMe(
	ctx context.Context,
	user *basedto.User,
	req *sessiondto.GetMeReq,
) (*sessiondto.GetMeResp, error) {
	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}
	if req.GetAccesses {
		loadOpts = append(loadOpts,
			bunex.SelectRelation("Accesses.ResourceProject"),
		)
	}

	dbUser, err := uc.userRepo.GetByID(ctx, uc.db, user.ID, loadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	userResp, err := sessiondto.TransformUserDetails(dbUser)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData := &sessiondto.GetMeDataResp{User: userResp}
	if user.Status == base.UserStatusPending && user.TotpSecret == "" {
		respData.NextStep = nextStepMfaSetup
	}

	return &sessiondto.GetMeResp{
		Data: respData,
	}, nil
}
