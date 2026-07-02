package appuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) ListApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.ListAppReq,
) (*appdto.ListAppResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	}

	if req.ParentID != "" {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.parent_id = ?", req.ParentID),
			bunex.SelectRelation("ParentApp",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
			bunex.SelectRelation("Settings",
				// NOTE: load http settings to extract active domain names of the app
				bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
			),
		)
	} else {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.parent_id IS NULL"),
		)
	}
	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.status IN (?)", bunex.List(req.Status)),
		)
	}
	if len(req.Env) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.env IN (?)", bunex.List(req.Env)),
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
	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.id IN (?)", bunex.List(auth.AllowObjectIDs)),
		)
	}

	apps, paging, err := uc.appRepo.List(ctx, uc.db, req.ProjectID, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats && len(apps) > 0 {
		serviceMap, err := uc.loadAppSwarmServices(ctx, apps[0].Project.Key, apps)
		if err != nil {
			return nil, apperrors.New(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApps(apps, transformationInput)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appdto.ListAppResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}

func (uc *UC) loadAppSwarmServices(
	ctx context.Context,
	projectKey string,
	apps []*entity.App,
) (map[string]*swarm.Service, error) {
	// Load all services of the project
	listResp, err := uc.dockerManager.ServiceListByStack(ctx, projectKey, func(opts *client.ServiceListOptions) {
		opts.Status = true
		if len(apps) == 1 && apps[0].ServiceID != "" {
			docker.FilterAdd(&opts.Filters, "id", apps[0].ServiceID)
		}
	})
	if err != nil {
		return nil, apperrors.New(err)
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

	resp := make(map[string]*swarm.Service, len(apps))
	for _, app := range apps {
		resp[app.ID] = serviceMap[app.ServiceID]
	}

	return resp, nil
}
