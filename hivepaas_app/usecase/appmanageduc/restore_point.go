package appmanageduc

import (
	"bytes"
	"context"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// markSlack is what a mark is given beyond its own wait: the exec's own start
// and end.
const markSlack = 30 * time.Second

// CreateRestorePoint names a moment of a managed database (design §4, "Restore
// points"): the engine makes the point and waits until the WAL it is in is
// archived, so the point is restorable as soon as it is listed. No license is
// asked: points are part of restoring, which is never gated.
func (uc *UC) CreateRestorePoint(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.CreateRestorePointReq,
) (*appmanageddto.CreateRestorePointResp, error) {
	resp := &appmanageddto.CreateRestorePointResp{}
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, point, err := uc.createRestorePoint(ctx, db, auth, req)
		if err != nil {
			return err
		}
		resp.Data = appmanageddto.TransformRestorePoint(point)
		return uc.recordPoint(ctx, db, auth, loaded, "managed-restore-point-create", auditdetail.New().
			Set("name", point.Name).Set("lsn", point.LSN).Set("kept", !point.KeepUntil.IsZero()))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (uc *UC) createRestorePoint(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appmanageddto.CreateRestorePointReq,
) (*loadedManaged, *entity.ManagedRestorePoint, error) {
	loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, nil, err
	}
	pg := loaded.managed.Postgres
	if pg == nil {
		return nil, nil, hperrors.NewNotFound("Managed archive")
	}
	if pg.RestorePoint(req.Name) != nil {
		return nil, nil, hperrors.NewAlreadyExist("Restore point").WithExtraDetail("%s", req.Name)
	}
	mark, err := loaded.engine.MarkRestorePoint(loaded.managed, req.Name)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	// When it was asked for: the point is after it, and the base backup a restore
	// to it fetches is the latest that starts before.
	at := timeutil.NowUTC()
	lsn, err := uc.runMark(ctx, db, loaded, mark)
	if err != nil {
		return nil, nil, err
	}
	point := &entity.ManagedRestorePoint{Name: req.Name, LSN: lsn, At: at, Note: req.Note}
	if auth != nil && auth.User != nil {
		point.By = auth.User.ID
	}
	if req.KeepUntil != nil {
		point.KeepUntil = req.KeepUntil.UTC()
	}
	pg.RestorePoints = append(pg.RestorePoints, point)
	return loaded, point, uc.saveManaged(ctx, db, loaded)
}

// runMark runs the engine's mark in the database's task, and reads the point's
// LSN from what it printed.
func (uc *UC) runMark(ctx context.Context, db database.Tx, loaded *loadedManaged,
	mark managedmodel.MarkCommand) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, mark.Timeout+markSlack)
	defer cancel()
	stdout := &bytes.Buffer{}
	resp, err := uc.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: gofn.Must(ulid.NewStringULID())},
			LogStore: tasklog.NewNullStore()},
		App:          loaded.app,
		Command:      &entity.CommandTemplate{Command: mark.Command},
		StdoutWriter: stdout,
	})
	if resp != nil && resp.ExitCode != nil && *resp.ExitCode == managedmodel.MarkExitNotArchived {
		return "", hperrors.Wrap(hperrors.ErrAppManagedRestorePointNotArchived)
	}
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if resp == nil || resp.ExitCode == nil || *resp.ExitCode != 0 {
		return "", hperrors.Wrap(hperrors.ErrActionFailed).WithExtraDetail("the restore point could not be made")
	}
	lsn, err := loaded.engine.MarkResult(stdout.String())
	return lsn, hperrors.Wrap(err)
}

// UpdateRestorePoint changes how long a point is kept, and its note.
func (uc *UC) UpdateRestorePoint(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UpdateRestorePointReq,
) (*appmanageddto.UpdateRestorePointResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, point, err := uc.updateRestorePoint(ctx, db, req)
		if err != nil {
			return err
		}
		return uc.recordPoint(ctx, db, auth, loaded, "managed-restore-point-update", auditdetail.New().
			Set("name", point.Name).Set("kept", !point.KeepUntil.IsZero()))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.UpdateRestorePointResp{}, nil
}

func (uc *UC) updateRestorePoint(ctx context.Context, db database.Tx, req *appmanageddto.UpdateRestorePointReq) (
	*loadedManaged, *entity.ManagedRestorePoint, error) {
	loaded, point, err := uc.loadPoint(ctx, db, req.ProjectID, req.AppID, req.Name)
	if err != nil {
		return nil, nil, err
	}
	point.Note = req.Note
	point.KeepUntil = time.Time{}
	if req.KeepUntil != nil {
		point.KeepUntil = req.KeepUntil.UTC()
	}
	return loaded, point, uc.saveManaged(ctx, db, loaded)
}

// DeleteRestorePoint removes a point's record. The mark stays in the WAL,
// harmlessly, and the archive is not touched: the retention it held back moves
// on at the next nightly run.
func (uc *UC) DeleteRestorePoint(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.DeleteRestorePointReq,
) (*appmanageddto.DeleteRestorePointResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, point, err := uc.deleteRestorePoint(ctx, db, req)
		if err != nil {
			return err
		}
		return uc.recordPoint(ctx, db, auth, loaded, "managed-restore-point-delete", auditdetail.New().
			Set("name", point.Name).Set("lsn", point.LSN))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.DeleteRestorePointResp{}, nil
}

func (uc *UC) deleteRestorePoint(ctx context.Context, db database.Tx, req *appmanageddto.DeleteRestorePointReq) (
	*loadedManaged, *entity.ManagedRestorePoint, error) {
	loaded, point, err := uc.loadPoint(ctx, db, req.ProjectID, req.AppID, req.Name)
	if err != nil {
		return nil, nil, err
	}
	pg := loaded.managed.Postgres
	pg.RestorePoints = slices.DeleteFunc(pg.RestorePoints, func(p *entity.ManagedRestorePoint) bool {
		return p == point
	})
	return loaded, point, uc.saveManaged(ctx, db, loaded)
}

func (uc *UC) loadPoint(ctx context.Context, db database.Tx, projectID, appID, name string) (
	*loadedManaged, *entity.ManagedRestorePoint, error) {
	loaded, err := uc.loadManaged(ctx, db, projectID, appID, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return nil, nil, err
	}
	if loaded.managed.Postgres == nil {
		return nil, nil, hperrors.NewNotFound("Restore point")
	}
	point := loaded.managed.Postgres.RestorePoint(name)
	if point == nil {
		return nil, nil, hperrors.NewNotFound("Restore point")
	}
	return loaded, point, nil
}

// saveManaged stores the engine's setting as it was changed.
func (uc *UC) saveManaged(ctx context.Context, db database.IDB, loaded *loadedManaged) error {
	if err := loaded.setting.SetData(loaded.managed); err != nil {
		return hperrors.Wrap(err)
	}
	loaded.setting.UpdateVer++
	loaded.setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(uc.settingRepo.Update(ctx, db, loaded.setting,
		bunex.UpdateColumns("data", "update_ver", "updated_at")))
}

func (uc *UC) recordPoint(ctx context.Context, db database.IDB, auth *basedto.Auth, loaded *loadedManaged,
	section string, detail *auditdetail.Builder) error {
	return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: loaded.app.ID,
		Source:   base.AuditLogSourceAPIAction,
		Section:  section,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    loaded.app.ID,
		ResName:  loaded.app.Name,
		Detail:   detail.String(),
	}))
}
