package appmanageduc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// restoredNameTime is the time a restored app's default name carries, to the
// second: two restores of one app in a minute must not take one name.
const restoredNameTime = "20060102-150405"

// RestoreManagedApp restores a managed database into a new app, to a target
// inside its restore window (design §4). The new app is made from the source's
// template binding - its template, version and parameters - through any
// creation's path, with its own data volume and its own prefix in the same
// storage; it is not the source's child, as it must outlive it. A task sees the
// restore through. The source is only read.
func (uc *UC) RestoreManagedApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.RestoreManagedReq,
) (*appmanageddto.RestoreManagedResp, error) {
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	pg := loaded.managed.Postgres
	if pg == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if loaded.app.ProjectEnvID != req.ProjectEnvID {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", req.AppID)
	}
	status, err := uc.statusOf(ctx, loaded)
	if err != nil {
		return nil, err
	}
	target, pointAt, err := resolveTarget(&req.Target, pg, status)
	if err != nil {
		return nil, err
	}
	binding, err := uc.templateBinding(ctx, loaded.app)
	if err != nil {
		return nil, err
	}
	tmpl, err := uc.appTemplateService.Template(ctx, binding.Template)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	create, err := restoreCreateReq(loaded, binding, tmpl.Template.Parameters, req, timeutil.NowUTC())
	if err != nil {
		return nil, err
	}
	source, err := uc.restoreSource(ctx, loaded, target, pointAt)
	if err != nil {
		return nil, err
	}

	created, task, err := uc.appTemplateUC.RestoreManagedApp(ctx, auth, create, source)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &appmanageddto.RestoreManagedResp{Data: &appmanageddto.RestoreManagedDataResp{
		App: created.Data.App, Task: &basedto.ObjectIDResp{ID: task.ID}}}
	err = auditservice.RecordAllowed(ctx, uc.auditService, uc.db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: loaded.app.ID,
		Source:   base.AuditLogSourceAPIAction,
		Section:  "managed-restore",
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    loaded.app.ID,
		ResName:  loaded.app.Name,
		Detail: auditdetail.New().Set("target", target.Kind).Set("restoredApp", created.Data.App.ID).
			String(),
	})
	return resp, hperrors.Wrap(err)
}

// resolveTarget checks a target against what the archive reaches - the restore
// window, the points recorded, the base backups kept - and resolves it: a point
// to its LSN, and when it was made.
func resolveTarget(
	req *appmanageddto.ManagedRestoreTargetReq,
	pg *entity.ManagedPostgres,
	status *managedmodel.Status,
) (entity.ManagedRestoreTarget, time.Time, error) {
	target := entity.ManagedRestoreTarget{Kind: req.Kind}
	window := status.RestoreWindow
	if window == nil {
		return target, time.Time{}, hperrors.Wrap(hperrors.ErrAppManagedRestoreNoBackup)
	}
	outside := func(at time.Time) error {
		return hperrors.Wrap(hperrors.ErrAppManagedRestoreOutsideWindow).
			WithParam("From", window.From.Format(time.RFC3339)).WithParam("To", window.To.Format(time.RFC3339)).
			ExposeParams("From", "To").WithExtraDetail("the target is %s", at.Format(time.RFC3339))
	}
	switch req.Kind {
	case entity.ManagedRestoreToLatest:
		return target, time.Time{}, nil
	case entity.ManagedRestoreToTime:
		at := req.Time.UTC()
		if at.Before(window.From) || at.After(window.To) {
			return target, time.Time{}, outside(at)
		}
		target.Time = at
		return target, time.Time{}, nil
	case entity.ManagedRestoreToPoint:
		point := pg.RestorePoint(req.Point)
		if point == nil {
			return target, time.Time{}, hperrors.NewNotFound("Restore point")
		}
		if point.At.Before(window.From) {
			return target, time.Time{}, outside(point.At)
		}
		target.Point, target.LSN = point.Name, point.LSN
		return target, point.At, nil
	case entity.ManagedRestoreToBackup:
		for _, backup := range status.Backups {
			if backup.Name == req.Backup {
				target.Backup = backup.Name
				return target, time.Time{}, nil
			}
		}
		return target, time.Time{}, hperrors.NewNotFound("Base backup")
	}
	return target, time.Time{}, hperrors.Wrap(hperrors.ErrAppManagedRestoreTargetInvalid)
}

// templateBinding is what the app was made from.
func (uc *UC) templateBinding(ctx context.Context, app *entity.App) (*entity.AppTemplateSettings, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, app.GetObjectScope(), base.SettingTypeAppTemplate, true,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.NewNotFound("App template")
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	binding, err := setting.AsAppTemplateSettings()
	return binding, hperrors.Wrap(err)
}

// restoreCreateReq is the request the new app is created with: the source's
// template, version, variant and image, its parameters with the new data
// volume, and its archive's storage and options. The prefix is the new app's,
// as any creation's is.
func restoreCreateReq(
	loaded *loadedManaged,
	binding *entity.AppTemplateSettings,
	defs []*templatemodel.Parameter,
	req *appmanageddto.RestoreManagedReq,
	now time.Time,
) (*apptemplatedto.CreateAppFromTemplateReq, error) {
	params, err := restoreParams(defs, binding.Params, req.DataVolume.ID)
	if err != nil {
		return nil, err
	}
	pg := loaded.managed.Postgres
	create := &apptemplatedto.CreateAppFromTemplateReq{
		ProjectID:    loaded.app.ProjectID,
		ProjectEnvID: loaded.app.ProjectEnvID,
		Name:         req.Name,
		Template:     binding.Template,
		Version:      binding.Version,
		Variant:      binding.Variant,
		ImageTag:     imageTag(binding.ImageOverride),
		Params:       params,
		Managed: &apptemplatedto.ManagedCreateReq{
			Storage: apptemplatedto.ManagedStorageReq{BackupRepo: objectIDReq(pg.Storage.BackupRepo),
				CloudStorage: objectIDReq(pg.Storage.CloudStorage), KeyAuth: objectIDReq(pg.Storage.KeyAuth)},
			ArchiveTimeout:      pg.ArchiveTimeout,
			RetainFull:          pg.RetainFull,
			BaseBackupSchedule:  loaded.cronOf(pg.Jobs.Nightly.ID),
			RestoreTestSchedule: loaded.cronOf(pg.Jobs.RestoreTest.ID),
			RestoreTestQuery:    pg.RestoreTest.CheckQuery,
		},
	}
	if create.Name == "" {
		create.Name = fmt.Sprintf("%s-restored-%s", loaded.app.Name, now.Format(restoredNameTime))
	}
	return create, nil
}

// restoreParams are the source's parameters, secrets decrypted, for the
// template's parameters as they are now - one the template no longer declares
// is left out, one it has added takes its default - and the data volume in place
// of the source's.
func restoreParams(defs []*templatemodel.Parameter, source map[string]*entity.AppTemplateParam,
	dataVolume string) (map[string]any, error) {
	params := map[string]any{}
	volumes := 0
	for _, def := range defs {
		if def.Type == templatemodel.ParamTypeVolume {
			volumes++
			params[def.Name] = dataVolume
			continue
		}
		param := source[def.Name]
		if param == nil {
			continue
		}
		if param.Secret.IsEmpty() {
			params[def.Name] = param.Value
			continue
		}
		value, err := param.Secret.GetPlain()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		params[def.Name] = value
	}
	if volumes != 1 {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).
			WithExtraDetail("a restore gives one data volume, and the template has %d", volumes)
	}
	return params, nil
}

// imageTag is an image reference's tag, empty for none.
func imageTag(image string) string {
	image, _, _ = strings.Cut(image, "@")
	slash := strings.LastIndex(image, "/")
	if colon := strings.LastIndex(image, ":"); colon > slash {
		return image[colon+1:]
	}
	return ""
}

func objectIDReq(id entity.ObjectID) *basedto.ObjectIDReq {
	if id.ID == "" {
		return nil
	}
	return &basedto.ObjectIDReq{ID: id.ID}
}

// cronOf is the schedule of one of the engine's jobs, empty for one gone.
func (l *loadedManaged) cronOf(jobID string) string {
	for _, setting := range l.jobs {
		if setting.ID != jobID {
			continue
		}
		if job, err := setting.AsSchedJob(); err == nil && job.Schedule != nil {
			return job.Schedule.CronExpr
		}
	}
	return ""
}

// restoreSource is the source's archive as the new app reads it - an S3 prefix
// with its credentials, or a volume repository's directory - its key, and the
// target.
func (uc *UC) restoreSource(
	ctx context.Context,
	loaded *loadedManaged,
	target entity.ManagedRestoreTarget,
	pointAt time.Time,
) (*managedmodel.RestoreSource, error) {
	pg := loaded.managed.Postgres
	key, err := pg.EncryptionKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	source := &managedmodel.RestoreSource{SourceAppID: loaded.app.ID, EncryptionKey: key, Target: target,
		PointAt: pointAt}
	scope := appScope(loaded.app)
	if pg.Storage.Placement == entity.ManagedPlacementOffSite {
		s3, resolveErr := uc.managedService.ResolveS3(ctx, uc.db, scope, pg.Storage)
		if resolveErr != nil {
			return nil, hperrors.Wrap(resolveErr)
		}
		source.Archive = &managedmodel.Archive{S3: s3}
		return source, nil
	}
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, scope, base.SettingTypeBackupRepo, pg.Storage.BackupRepo.ID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	repo, err := setting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	source.Archive = &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{VolumeID: repo.Volume.ID,
		Subpath: pg.Storage.Prefix}}
	return source, nil
}
