package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
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
		return nil, apperrors.New(err)
	}

	userResp, err := sessiondto.TransformUserDetails(dbUser)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData := &sessiondto.GetMeDataResp{User: userResp}

	if config.Current.SystemInfo.NextStep != "" && user.IsAdmin() {
		sysStatus, err := uc.systemStatusRepo.Get(ctx, uc.db)
		if err != nil {
			return nil, apperrors.New(err)
		}
		config.Current.SystemInfo.NextStep = sysStatus.NextStep
		respData.NextStep = string(sysStatus.NextStep)
	}

	if user.Status == base.UserStatusPending && user.TotpSecret == "" {
		respData.NextStep = nextStepMfaSetup
	}

	return &sessiondto.GetMeResp{
		Data: respData,
	}, nil
}
