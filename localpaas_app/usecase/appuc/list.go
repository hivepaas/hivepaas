package appuc

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) ListApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.ListAppReq,
) (*appdto.ListAppResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Project"),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	}

	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.status IN (?)", bunex.List(req.Status)),
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
		return nil, apperrors.Wrap(err)
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats && len(apps) > 0 {
		serviceMap, err := uc.loadAppSwarmServices(ctx, apps[0].Project.Key, apps)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApps(apps, transformationInput)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
	// TODO: implement caching?

	// Special case: only one app
	if len(apps) == 1 {
		app := apps[0]
		if app.ServiceID == "" {
			return nil, nil
		}
		service, err := uc.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return map[string]*swarm.Service{app.ID: service}, nil
	}

	// Load all services of the project
	services, err := uc.dockerManager.ServiceListByStack(ctx, projectKey, func(opts *swarm.ServiceListOptions) {
		opts.Status = true
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	serviceMap := make(map[string]*swarm.Service, len(services))
	for _, service := range services {
		serviceMap[service.ID] = &service
	}

	resp := make(map[string]*swarm.Service, len(apps))
	for _, app := range apps {
		resp[app.ID] = serviceMap[app.ServiceID]
	}

	return resp, nil
}
