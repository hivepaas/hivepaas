package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
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
		// Must match useruc.GetUser: permissions are granted per project env, so
		// loading only the project relation would hide every env-level grant.
		loadOpts = append(loadOpts,
			bunex.SelectRelation("Accesses.ResourceProject.ProjectEnvs"),
			bunex.SelectRelation("Accesses.ResourceProjectEnv.Project.ProjectEnvs"),
		)
	}

	dbUser, err := uc.userRepo.GetByID(ctx, uc.db, user.ID, loadOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	userResp, err := sessiondto.TransformUserDetails(dbUser)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData := &sessiondto.GetMeDataResp{User: userResp}

	if config.CurrentSystemInfo().NextStep != "" && user.IsAdmin() {
		sysStatus, err := uc.systemStatusRepo.Get(ctx, uc.db)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		config.SetInstallationStep(sysStatus.NextStep)
		respData.NextStep = string(sysStatus.NextStep)
		if err = uc.addSetupChecklist(ctx, user, respData); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if user.Status == base.UserStatusPending && user.TotpSecret == "" {
		respData.NextStep = nextStepMfaSetup
	}

	return &sessiondto.GetMeResp{
		Data: respData,
	}, nil
}

// addSetupChecklist gives an admin what the installation still has to do, while
// the step says so. Finding all of it done clears the step, so the card goes
// for every admin without anyone closing it.
func (uc *UC) addSetupChecklist(
	ctx context.Context,
	user *basedto.User,
	respData *sessiondto.GetMeDataResp,
) error {
	if respData.NextStep != base.InstallationStepGetStarted {
		return nil
	}
	checklist, err := uc.getStartedService.Checklist(ctx, uc.db, user.TotpSecret != "")
	if err != nil {
		return hperrors.Wrap(err)
	}
	if checklist.AllDone() {
		if err = uc.getStartedService.Finish(ctx, uc.db); err != nil {
			return hperrors.Wrap(err)
		}
		respData.NextStep = ""
		return nil
	}
	respData.SetupChecklist = getstarteddto.TransformChecklist(checklist)
	return nil
}
