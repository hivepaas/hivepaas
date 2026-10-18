package apptemplateuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

const (
	// restoreTaskMaxRetry and restoreTaskRetryDelay retry a step that failed on
	// its way - docker, the database: each step can run again.
	restoreTaskMaxRetry   = 5
	restoreTaskRetryDelay = time.Minute
)

// managedRestore makes a creation the restore of a managed database: what the
// engine recovers the new app from, and the task made with it to see the
// restore through. Nil is an ordinary creation.
type managedRestore struct {
	source *managedmodel.RestoreSource
	task   *entity.Task
}

// RestoreManagedApp creates a managed database restored from another, from the
// template binding the caller made req of. It is any creation's path, but the
// license is not asked - restoring is never gated - the engine is given what the
// app recovers from, and the task that sees the restore through is made with
// the app, and scheduled once it is committed. It returns the task.
func (uc *UC) RestoreManagedApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	source *managedmodel.RestoreSource,
) (*apptemplatedto.CreateAppFromTemplateResp, *entity.Task, error) {
	if req.Managed == nil || source == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("a restore needs its archive and source")
	}
	restore := &managedRestore{source: source}
	resp, err := uc.createFromTemplate(ctx, auth, req, restore)
	if err != nil {
		return nil, nil, err
	}
	return resp, restore.task, nil
}

// insertRestoreTask inserts, for a restore, the task that sees it through, on
// the app the request is about.
func (uc *UC) insertRestoreTask(
	ctx context.Context,
	db database.IDB,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
	created []*appprovisionservice.ProvisionAppResp,
	restore *managedRestore,
) error {
	if restore == nil {
		return nil
	}
	var app *entity.App
	for i, target := range apps {
		if target == mainApp(apps) && i < len(created) {
			app = created[i].App
		}
	}
	if app == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("a restore created no app to restore into")
	}
	restore.task = newRestoreTask(app, req.ProjectID, restore.source.SourceAppID, timeutil.NowUTC())
	return hperrors.Wrap(uc.taskRepo.Insert(ctx, db, restore.task))
}

func newRestoreTask(app *entity.App, projectID, sourceAppID string, now time.Time) *entity.Task {
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		TargetID: app.ID,
		Type:     base.TaskTypeManagedRestore,
		Status:   base.TaskStatusNotStarted,
		Version:  entity.CurrentTaskVersion,
		Config: entity.TaskConfig{MaxRetry: restoreTaskMaxRetry,
			RetryDelay: timeutil.Duration(restoreTaskRetryDelay)},
		RunAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task.MustSetArgs(&entity.TaskManagedRestoreArgs{ProjectID: projectID, AppID: app.ID, SourceAppID: sourceAppID})
	task.MustSetOutput(&entity.TaskManagedRestoreState{Step: entity.ManagedRestoreStepRecover})
	return task
}
