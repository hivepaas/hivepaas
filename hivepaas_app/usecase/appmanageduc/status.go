package appmanageduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// GetManagedStatus is a managed app's state as its engine reads it from the last
// run of each of its jobs. Nothing asks the database: what the health check, the
// nightly backup and the restore test last wrote is what there is to say.
func (uc *UC) GetManagedStatus(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmanageddto.GetManagedStatusReq,
) (*appmanageddto.GetManagedStatusResp, error) {
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	runs := make(map[string]*managedmodel.JobRun, len(loaded.jobs))
	for _, job := range loaded.jobs {
		tasks, _, err := uc.taskRepo.ListByTarget(ctx, uc.db, job.ID, nil,
			bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusDone, base.TaskStatusFailed),
			bunex.SelectOrder("task.ended_at DESC"),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if len(tasks) == 0 {
			continue
		}
		run := &managedmodel.JobRun{Status: tasks[0].Status, EndedAt: tasks[0].EndedAt}
		if output, outputErr := tasks[0].OutputAsSchedJobOutputs(); outputErr == nil && output != nil {
			run.Outputs = output.Outputs
		}
		runs[job.ID] = run
	}
	status, err := loaded.engine.Status(&managedmodel.StatusReq{
		Managed: loaded.managed,
		Runs:    runs,
		Now:     timeutil.NowUTC(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.GetManagedStatusResp{
		Data: appmanageddto.TransformManagedStatus(loaded.managed.Engine, loaded.managed.EngineAPI, status),
	}, nil
}
