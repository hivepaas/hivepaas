package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) ListTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.ListTaskReq,
) (*taskdto.ListTaskResp, error) {
	if req.Scope == nil {
		req.Scope = entity.NewObjectScopeGlobal()
	}
	targetScope := req.Scope
	switch {
	case req.ProjectID != "":
		targetScope = entity.NewObjectScopeProject(req.ProjectID)
	case req.ProjectEnvID != "":
		projectID, env := projecthelper.ParseProjectEnvID(req.ProjectEnvID)
		if projectID != "" && env != "" {
			targetScope = entity.NewObjectScopeProjectEnv(projectID, env)
		}
	case req.AppID != "":
		targetScope = entity.NewObjectScopeApp(req.AppID, "", "", "")
	}
	targetScope.NoInherited = req.Scope.NoInherited
	if req.ScopeOnly {
		targetScope.NoInherited = true
	}

	listResp, err := uc.taskService.ListTask(ctx, uc.db, &taskservice.ListTaskReq{
		Scope:     targetScope,
		Types:     req.Type,
		TargetIDs: req.TargetID,
		Statuses:  req.Status,
		FromDate:  req.FromDate,
		ToDate:    req.ToDate,
		Search:    req.Search,
		Paging:    req.Paging,
		ExtraSelectOpts: []bunex.SelectQueryOption{
			bunex.SelectRelation("TargetJob",
				bunex.SelectColumns("id", "type", "kind", "name", "status"),
			),
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	refObjects.AddObjectScope(targetScope)
	err = uc.loadTaskRefData(ctx, uc.db, listResp.Tasks, &refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTasks(listResp.Tasks, listResp.TaskInfoMap, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.ListTaskResp{
		Meta: &basedto.ListMeta{Page: listResp.PagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) loadTaskRefData(
	ctx context.Context,
	db database.IDB,
	tasks []*entity.Task,
	refObjects **entity.RefObjects,
) error {
	if len(tasks) == 0 {
		return nil
	}
	refIDs := &entity.RefObjectIDs{}
	for _, task := range tasks {
		if task == nil {
			continue
		}
		refIDs.AddRefIDs(task.GetRefObjectIDs())
	}
	err := uc.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, refObjects, nil, false, refIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
