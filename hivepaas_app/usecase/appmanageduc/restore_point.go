package appmanageduc

import (
	"context"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// CreateRestorePoint names a moment of a managed database (design §4, "Restore
// points"): the engine makes the point and waits until the WAL it is in is
// archived, so the point is restorable as soon as it is listed. No license is
// asked: points are part of restoring, which is never gated. It answers the
// point as Status lists it, with until when it can be restored.
func (uc *UC) CreateRestorePoint(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.CreateRestorePointReq,
) (*appmanageddto.CreateRestorePointResp, error) {
	marked, err := uc.managedService.MarkRestorePoint(ctx, markPointReq(auth, req,
		func(ctx context.Context, db database.IDB, app *entity.App, point *entity.ManagedRestorePoint) error {
			return uc.recordPoint(ctx, db, auth, app, "managed-restore-point-create", auditdetail.New().
				Set("name", point.Name).Set("lsn", point.LSN).Set("kept", !point.KeepUntil.IsZero()))
		}))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &appmanageddto.CreateRestorePointResp{Data: appmanageddto.TransformRestorePoint(marked.Point)}
	// Until when it can be restored is read as Status reads it. The point is made
	// whether or not that can be read.
	if loaded, loadErr := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID); loadErr == nil {
		if status, statusErr := uc.statusOf(ctx, loaded); statusErr == nil {
			for _, point := range status.RestorePoints {
				if point.Point.Name == marked.Point.Name {
					resp.Data = appmanageddto.TransformRestorePointStatus(point)
				}
			}
		}
	}
	return resp, nil
}

// markPointReq is a person's request for a point, as the engine's service takes
// it.
func markPointReq(auth *basedto.Auth, req *appmanageddto.CreateRestorePointReq,
	record func(context.Context, database.IDB, *entity.App, *entity.ManagedRestorePoint) error,
) *managedservice.MarkPointReq {
	out := &managedservice.MarkPointReq{ProjectID: req.ProjectID, AppID: req.AppID, Name: req.Name, Note: req.Note,
		Record: record}
	if auth != nil && auth.User != nil {
		out.By = auth.User.ID
	}
	if req.KeepUntil != nil {
		out.KeepUntil = req.KeepUntil.UTC()
	}
	return out
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
		return uc.recordPoint(ctx, db, auth, loaded.app, "managed-restore-point-update", auditdetail.New().
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
		return uc.recordPoint(ctx, db, auth, loaded.app, "managed-restore-point-delete", auditdetail.New().
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

func (uc *UC) recordPoint(ctx context.Context, db database.IDB, auth *basedto.Auth, app *entity.App,
	section string, detail *auditdetail.Builder) error {
	return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIAction,
		Section:  section,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.String(),
	}))
}
