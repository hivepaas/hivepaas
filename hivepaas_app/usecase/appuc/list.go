package appuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

//nolint:gocognit,funlen
func (uc *UC) ListApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.ListAppReq,
) (*appdto.ListAppResp, error) {
	// When parent ID is passed, user wants to list preview apps of an app.
	// We need to verify the app preview feature is enabled.
	if req.ParentID != "" {
		_, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID,
			req.ParentID, false, false)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if featureSettings.PreviewSettings != nil && !featureSettings.PreviewSettings.Enabled {
			return nil, hperrors.Wrap(hperrors.ErrFeatureDisabled).WithParam("Name", "app preview")
		}
	}

	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	}

	if req.ParentID != "" {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.parent_id = ?", req.ParentID),
			bunex.SelectRelation("ParentApp",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
			bunex.SelectRelation("Settings",
				// NOTE: load routing settings to extract active domain names of the app,
				// and the kind to say what each app runs.
				bunex.SelectWhereIn("setting.type IN (?)",
					base.SettingTypeAppRouting, base.SettingTypeAppKind),
			),
		)
	} else {
		// The apps of an environment are the ones somebody made, not the ones made
		// underneath them. Whichever way a child belongs to its owner, it is shown
		// beside that owner instead - see attachChildApps.
		listOpts = append(listOpts, excludeChildApps()...)
		listOpts = append(listOpts,
			// The kind alone here: what an app runs is what a list is filtered by,
			// and loading routing settings as well would change what every other
			// caller of this list receives.
			bunex.SelectRelation("Settings",
				bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
			),
		)
		if req.GetChildApps {
			listOpts = append(listOpts,
				bunex.SelectRelation("ChildApps",
					bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...)),
				bunex.SelectRelation("LogicalChildApps",
					bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...)),
			)
		}
	}
	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhereIn("app.status IN (?)", req.Status...),
		)
	}

	// Filter by search keyword
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("app.name ILIKE ?", keyword),
				bunex.SelectWhereOr("app.note ILIKE ?", keyword),
			),
		)
	}

	if req.ProjectEnvID != "" {
		listOpts = append(listOpts,
			bunex.SelectJoin("JOIN project_envs ON project_envs.id = app.project_env_id"),
			bunex.SelectWhere("project_envs.id = ?", req.ProjectEnvID),
		)
	}

	allowedAllIDs, allowedIDs := auth.AllowedApps(nil)
	if !allowedAllIDs {
		if len(allowedIDs) == 0 { // return empty result
			return &appdto.ListAppResp{Meta: basedto.NewEmptyListMeta()}, nil
		}
		listOpts = append(listOpts,
			bunex.SelectWhereIn("app.id IN (?)", allowedIDs...),
		)
	}

	apps, paging, err := uc.appRepo.List(ctx, uc.db, req.ProjectID, &req.Paging, listOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// NOTE: make sure we init the project env and project for the parent app and the child apps
	for _, app := range apps {
		if app.ParentApp != nil {
			app.ParentApp.Project = app.Project
			app.ParentApp.ProjectEnv = app.ProjectEnv
		}
		for _, childApp := range gofn.Concat(app.ChildApps, app.LogicalChildApps) {
			if childApp.Project == nil {
				childApp.Project = app.Project
			}
			if childApp.ProjectEnv == nil {
				childApp.ProjectEnv = app.ProjectEnv
			}
		}
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats && len(apps) > 0 {
		serviceMap, err := uc.loadAppSwarmServices(ctx, apps[0].Project.Key, apps)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApps(apps, transformationInput)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdto.ListAppResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}

// An app can belong to another in two ways, and a listing shows neither of them
// beside the app they belong to:
//
//   - a child app names its parent in app.parent_id, which is what a preview
//     deployment is: a copy of the app it was made from;
//   - a logical child app names it in app.logical_parent_id, which says it was
//     created to serve that app. A template's dependencies are these, and so are
//     the databases a preview clones for itself.
//
// Both are left out unless they are asked for, because a person looking at the
// apps of an environment is looking for the ones they made, not the six a
// template created underneath them.
func excludeChildApps() []bunex.SelectQueryOption {
	return []bunex.SelectQueryOption{
		bunex.SelectWhere("app.parent_id IS NULL"),
		bunex.SelectWhere("app.logical_parent_id IS NULL"),
	}
}

func (uc *UC) loadAppSwarmServices(
	ctx context.Context,
	projectKey string,
	apps []*entity.App,
) (map[string]*swarm.Service, error) {
	allApps := make([]*entity.App, 0, len(apps)*2) //nolint:mnd
	for _, app := range apps {
		allApps = append(allApps, app)
		allApps = append(allApps, app.ChildApps...)
		allApps = append(allApps, app.LogicalChildApps...)
	}
	// Load all services of the project
	listResp, err := uc.dockerManager.ServiceListByStack(ctx, projectKey, func(opts *client.ServiceListOptions) {
		opts.Status = true
		if len(allApps) == 1 && allApps[0].ServiceID != "" {
			docker.FilterAdd(&opts.Filters, "id", allApps[0].ServiceID)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	services := listResp.Items
	serviceMap := make(map[string]*swarm.Service, len(services))
	for i := range services {
		svc := &services[i]
		serviceMap[svc.ID] = svc

		// NOTE: If no `task status` returned, assign 0 to avoid no data returned to client
		if svc.ServiceStatus == nil {
			svc.ServiceStatus = &swarm.ServiceStatus{}
		}
	}

	resp := make(map[string]*swarm.Service, len(allApps))
	for _, app := range allApps {
		resp[app.ID] = serviceMap[app.ServiceID]
	}

	return resp, nil
}
