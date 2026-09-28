package backupsnapshotuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
	"github.com/hivepaas/hivepaas/services/backup"
)

// RestoreBackupSnapshot records a task that puts a snapshot the view reaches
// back into an app the caller may change.
func (uc *UC) RestoreBackupSnapshot(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.RestoreBackupSnapshotReq,
) (*backupsnapshotdto.RestoreBackupSnapshotResp, error) {
	var task *entity.Task
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
			return hperrors.Wrap(err)
		}
		record, repos, err := uc.findInReach(ctx, db, auth, req.Scope, req.ID, base.ActionTypeRead)
		if err != nil {
			return hperrors.Wrap(err)
		}
		target, snapshot, err := uc.snapshotTarget(ctx, db, record, repos)
		if err != nil {
			return hperrors.Wrap(err)
		}
		kind, err := uc.kindOf(ctx, db, record, repos, target, snapshot)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = restoreFits(kind, req); err != nil {
			return hperrors.Wrap(err)
		}
		app, err := uc.restoreTargetApp(ctx, db, auth, req)
		if err != nil {
			return hperrors.Wrap(err)
		}

		args := &entity.TaskBackupRestoreArgs{
			ProjectID: app.ProjectID, AppID: app.ID, RepoID: target.RepoSetting.ID, SnapshotID: snapshot.ID,
		}
		if kind.command {
			args.Command, args.FileName = req.Command.ToEntity(), kind.fileName
		} else {
			args.Volume = entity.ObjectID{ID: req.Volume.ID}
			args.Subpath, args.SnapshotPath = req.Subpath, req.SnapshotPath
			args.StopApp, args.Mode = req.StopApp, req.Mode
		}
		if err = uc.recordRestore(ctx, db, auth, app, snapshot, target.RepoSetting, args); err != nil {
			return hperrors.Wrap(err)
		}
		task = restoreTask(record.ID, app, args, timeutil.NowUTC())
		return hperrors.Wrap(uc.taskRepo.Insert(ctx, db, task))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupsnapshotdto.RestoreBackupSnapshotResp{
		Data: &backupsnapshotdto.RestoreBackupSnapshotDataResp{Task: &basedto.ObjectIDResp{ID: task.ID}},
	}, nil
}

// restoreKind is what a snapshot is restored as: a command's file, or a
// volume's directory.
type restoreKind struct {
	command  bool
	fileName string
}

// kindOf is the snapshot's kind, from its tags and its job, or its content.
func (uc *UC) kindOf(
	ctx context.Context,
	db database.IDB,
	record *entity.Setting,
	repos []*entity.Setting,
	target *backupreposervice.RepoTarget,
	snapshot *entity.BackupSnapshot,
) (*restoreKind, error) {
	tags, err := uc.tagsOf(ctx, db, []*entity.Setting{record})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refs, err := uc.snapshotRefs(ctx, db, repos, tags)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	named := backupsnapshotdto.TransformBackupSnapshot(record, tags[record.ID], refs)
	fileName := ""
	if named.Job != nil {
		fileName = named.Job.FileName
	}
	kind, err := snapshotKind(named.Source, fileName, func() ([]backup.SnapshotEntry, error) {
		entries, err := uc.backupRepoService.ListEntries(ctx, db, &backupreposervice.ListEntriesReq{
			RepoTarget: *target, SnapshotID: snapshot.ID,
		})
		return entries, hperrors.Wrap(err)
	})
	return kind, hperrors.Wrap(err)
}

// snapshotKind is a snapshot's kind: its source's, as its tag or its job says.
// One whose source is not known - taken outside HivePaaS, or before the tag -
// goes by what it holds: one file is a command's.
func snapshotKind(
	source base.SchedJobDataBackupSource,
	fileName string,
	listRoot func() ([]backup.SnapshotEntry, error),
) (*restoreKind, error) {
	if source == base.SchedJobDataBackupSourceVolume {
		return &restoreKind{}, nil
	}
	if source == base.SchedJobDataBackupSourceCommand && fileName != "" {
		return &restoreKind{command: true, fileName: fileName}, nil
	}
	root, err := listRoot()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	oneFile := len(root) == 1 && !root[0].Dir
	switch {
	case oneFile:
		return &restoreKind{command: true, fileName: root[0].Name}, nil
	case source == base.SchedJobDataBackupSourceCommand:
		return nil, hperrors.NewArgumentInvalid("snapshot").
			WithExtraDetail("a command's snapshot holds one file, and this one does not")
	}
	return &restoreKind{}, nil
}

// restoreFits refuses a restore of the other kind than the snapshot's.
func restoreFits(kind *restoreKind, req *backupsnapshotdto.RestoreBackupSnapshotReq) error {
	if kind.command && req.Command == nil {
		return hperrors.NewArgumentInvalid("command").
			WithExtraDetail("the snapshot is a command's output: it is restored through a command")
	}
	if !kind.command && req.Volume.ID == "" {
		return hperrors.NewArgumentInvalid("volume").
			WithExtraDetail("the snapshot is a directory: it is restored into a volume")
	}
	return nil
}

// restoreTargetApp is the app the data goes into: one the caller may change,
// that mounts the volume as its own directory for a volume's restore, and that
// has no restore in flight.
func (uc *UC) restoreTargetApp(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *backupsnapshotdto.RestoreBackupSnapshotReq,
) (*entity.App, error) {
	app, err := uc.appRepo.GetByID(ctx, db, "", req.TargetApp.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_, env := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
	allowed, err := uc.PermissionManager.NewVisibility(db, auth).AllowsProjectEnv(ctx, app.ProjectID, env,
		base.ActionTypeWrite)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !allowed {
		return nil, hperrors.Wrap(hperrors.ErrForbidden).WithExtraDetail("you may not change the target app")
	}
	if req.Volume.ID != "" {
		if err = uc.dataBackupService.CheckAppVolume(ctx, db, app, req.Volume.ID); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	inFlight, _, err := uc.taskRepo.List(ctx, db, nil, nil, restoreInFlightOpts(app.ID)...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(inFlight) > 0 {
		return nil, hperrors.NewArgumentInvalid("targetApp").
			WithExtraDetail("a restore into this app has not ended yet: task %s", inFlight[0].ID)
	}
	return app, nil
}

// restoreInFlightOpts find the restores into an app that have not ended.
func restoreInFlightOpts(appID string) []bunex.SelectQueryOption {
	return []bunex.SelectQueryOption{
		bunex.SelectWhere("task.type = ?", base.TaskTypeBackupRestore),
		bunex.SelectWhere("task.object_id = ?", appID),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectLimit(1),
	}
}

// restoreTask is the task of a restore into app.
func restoreTask(recordID string, app *entity.App, args *entity.TaskBackupRestoreArgs, now time.Time) *entity.Task {
	task := &entity.Task{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		TargetID:  recordID,
		Type:      base.TaskTypeBackupRestore,
		Status:    base.TaskStatusNotStarted,
		Version:   entity.CurrentTaskVersion,
		RunAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task.MustSetArgs(args)
	return task
}

// recordRestore records who asked for a restore into an app, and of what.
func (uc *UC) recordRestore(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	snapshot *entity.BackupSnapshot,
	repo *entity.Setting,
	args *entity.TaskBackupRestoreArgs,
) error {
	detail := auditdetail.New().Set("snapshot", snapshot.ShortID).Set("repository", repo.Name)
	if args.Command == nil {
		detail.Set("mode", string(args.Mode)).Set("stopApp", args.StopApp)
	}
	err := auditservice.RecordAllowed(ctx, uc.AuditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIAction,
		Section:  "backup-restore",
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.String(),
	})
	return hperrors.Wrap(err)
}
