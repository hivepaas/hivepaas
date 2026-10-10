package fileuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filecodec"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// LoadDataFile records a task that feeds an app's data file to a command run
// in the app, on its stdin: a dump a job saved, loaded back into the app's
// database. One load into an app runs at a time.
func (uc *UC) LoadDataFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.LoadDataFileReq,
) (*filedto.LoadDataFileResp, error) {
	var task *entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		file, err := uc.fileRepo.GetByID(ctx, db, req.ID,
			bunex.SelectWhere("file.status = ?", base.FileStatusActive),
			bunex.SelectWhere("file.object_id = ?", req.AppID),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if file.Deleted {
			return hperrors.NewNotFound("File")
		}
		if filecodec.Layers(file.Name).Encrypted && req.Passphrase == "" {
			return hperrors.NewArgumentInvalid("passphrase").
				WithExtraDetail("the file was saved encrypted: its passphrase reads it")
		}
		app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
			bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
			bunex.SelectRelation("ProjectEnv"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.noLoadInFlight(ctx, db, app.ID); err != nil {
			return hperrors.Wrap(err)
		}
		err = auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: app.ID,
			Source:   base.AuditLogSourceAPIAction,
			Section:  "data-file-load",
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    app.ID,
			ResName:  app.Name,
			Detail:   auditdetail.New().Set("file", file.Name).String(),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		task = loadTask(file, app, &entity.TaskDataFileLoadArgs{
			ProjectID:  app.ProjectID,
			AppID:      app.ID,
			Command:    req.Command.ToEntity(),
			Passphrase: entity.NewEncryptedField(req.Passphrase),
		})
		return hperrors.Wrap(uc.taskRepo.Insert(ctx, db, task))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &filedto.LoadDataFileResp{
		Data: &filedto.LoadDataFileDataResp{Task: &basedto.ObjectIDResp{ID: task.ID}},
	}, nil
}

// noLoadInFlight refuses a load into an app while one has not ended: two
// feeding one database at once leave it as neither would.
func (uc *UC) noLoadInFlight(ctx context.Context, db database.IDB, appID string) error {
	inFlight, _, err := uc.taskRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeDataFileLoad),
		bunex.SelectWhere("task.object_id = ?", appID),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(inFlight) > 0 {
		return hperrors.NewArgumentInvalid("app").
			WithExtraDetail("a load into this app has not ended yet: task %s", inFlight[0].ID)
	}
	return nil
}

// loadTask is the task of a load of the file into the app.
func loadTask(file *entity.File, app *entity.App, args *entity.TaskDataFileLoadArgs) *entity.Task {
	now := timeutil.NowUTC()
	task := &entity.Task{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		TargetID:  file.ID,
		Type:      base.TaskTypeDataFileLoad,
		Status:    base.TaskStatusNotStarted,
		Version:   entity.CurrentTaskVersion,
		RunAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task.MustSetArgs(args)
	return task
}
