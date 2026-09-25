package appsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envlink"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ListEnvLinkTargets(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.ListEnvLinkTargetsReq,
) (*appsettingsdto.ListEnvLinkTargetsResp, error) {
	self, apps, kinds, err := uc.loadEnvLinkApps(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.ListEnvLinkTargetsResp{
		Data: gofn.MapSlice(envlink.Targets(self, apps, kinds), appsettingsdto.TransformEnvLinkTarget),
	}, nil
}

func (uc *UC) GetEnvLinkSuggestions(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetEnvLinkSuggestionsReq,
) (*appsettingsdto.GetEnvLinkSuggestionsResp, error) {
	self, apps, kinds, err := uc.loadEnvLinkApps(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// A target is one Targets offers: of the env, not the app, not a preview.
	target, found := gofn.Find(envlink.Targets(self, apps, kinds), func(t *envlink.Target) bool {
		return t.ID == req.TargetAppID
	})
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", req.TargetAppID)
	}
	targetApp, _ := gofn.Find(apps, func(a *entity.App) bool { return a.ID == target.ID })

	// Read, never returned: only whether the port and the password are set.
	shared, err := uc.envVarService.BuildSharedEnvVarsInApp(ctx, uc.db, targetApp, envvarservice.EnvBuildOptions{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.GetEnvLinkSuggestionsResp{Data: &appsettingsdto.EnvLinkSuggestionsResp{
		Target: appsettingsdto.TransformEnvLinkTarget(target),
		Groups: appsettingsdto.TransformEnvLinkGroups(envlink.Suggest(target, envlink.FactsOf(shared))),
	}}, nil
}

// loadEnvLinkApps is the app, the apps of its env with what building their
// environment needs, and their kinds by app id.
func (uc *UC) loadEnvLinkApps(
	ctx context.Context, projectID, appID string,
) (*entity.App, []*entity.App, map[string]*entity.AppKindSettings, error) {
	self, err := uc.appService.LoadApp(ctx, uc.db, projectID, appID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	apps, _, err := uc.appRepo.List(ctx, uc.db, projectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectWhere("app.project_env_id = ?", self.ProjectEnvID),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	kinds := make(map[string]*entity.AppKindSettings, len(apps))
	if len(apps) == 0 {
		return self, apps, kinds, nil
	}
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.object_id IN (?)", gofn.MapSlice(apps, func(a *entity.App) string { return a.ID })...),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	for _, setting := range settings {
		kind, err := setting.AsAppKindSettings()
		if err != nil {
			return nil, nil, nil, hperrors.Wrap(err)
		}
		kinds[setting.ObjectID] = kind
	}
	return self, apps, kinds, nil
}
