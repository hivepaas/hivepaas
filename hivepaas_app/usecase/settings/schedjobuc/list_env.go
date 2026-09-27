package schedjobuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ListEnvSchedJob lists a project env's scheduled jobs and those of the apps in
// it, each saying whose it is: the list behind an env's scheduled jobs page and
// the picker of an env sequence's steps.
func (uc *UC) ListEnvSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ListEnvSchedJobReq,
) (*schedjobdto.ListEnvSchedJobResp, error) {
	req.Type = currentSettingType
	req.Scope.IncludeEnvApps = true
	for _, jobType := range req.JobTypes {
		req.Kinds = append(req.Kinds, string(jobType))
	}
	var extraOpts []bunex.SelectQueryOption
	if req.AppID != "" {
		extraOpts = append(extraOpts, bunex.SelectWhere("setting.object_id = ?", req.AppID))
	}
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{
		ExtraLoadOpts: extraOpts,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if resp.RefObjects == nil {
		resp.RefObjects = entity.NewRefObjects()
	}
	if err = uc.loadSequenceMembers(ctx, uc.DB, resp.RefObjects, resp.Data...); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.loadOwnerApps(ctx, resp.RefObjects, resp.Data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := schedjobdto.TransformEnvSchedJobs(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &schedjobdto.ListEnvSchedJobResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}

// loadOwnerApps adds to refObjects the apps the app jobs among jobSettings
// belong to.
func (uc *UC) loadOwnerApps(ctx context.Context, refObjects *entity.RefObjects, jobSettings []*entity.Setting) error {
	var appIDs []string
	for _, setting := range jobSettings {
		if setting.Scope == base.ObjectScopeApp {
			appIDs = append(appIDs, setting.ObjectID)
		}
	}
	if len(appIDs) == 0 {
		return nil
	}
	err := uc.SettingService.LoadRefObjectsByIDsSkipMissing(ctx, uc.DB, &refObjects, nil, false,
		&entity.RefObjectIDs{RefAppIDs: gofn.ToSet(appIDs)})
	return hperrors.Wrap(err)
}
