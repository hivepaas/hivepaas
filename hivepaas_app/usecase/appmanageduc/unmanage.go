package appmanageduc

import (
	"context"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// UnmanageApp makes a managed app an ordinary one: its engine's setting and the
// jobs it made go, and with them the locks. Nothing the database needs goes -
// its artifacts, secrets, scripts and arguments stay, so it goes on archiving
// where it did, and the archive stays where it is for the person to keep or
// delete. Base backups, retention, the health check and the restore test stop:
// they are the person's from now on.
func (uc *UC) UnmanageApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UnmanageAppReq,
) (*appmanageddto.UnmanageAppResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectFor("UPDATE OF app"),
			bunex.SelectRelation("Settings",
				bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppManaged, base.SettingTypeSchedJob),
				bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
			),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		deleting, err := unmanagedSettings(app, timeutil.NowUTC())
		if err != nil {
			return err
		}
		for _, setting := range deleting {
			if err = uc.settingRepo.Update(ctx, db, setting, bunex.UpdateColumns("deleted_at", "update_ver")); err != nil {
				return hperrors.Wrap(err)
			}
			if setting.Type == base.SettingTypeSchedJob {
				// Its coming runs go with it, as when a person deletes a job.
				if err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, true); err != nil {
					return hperrors.Wrap(err)
				}
			}
		}
		return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: app.ID,
			Source:   base.AuditLogSourceAPIAction,
			Section:  "unmanage",
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    app.ID,
			ResName:  app.Name,
			Detail:   auditdetail.New().Set("removedSettings", len(deleting)).String(),
		}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.UnmanageAppResp{}, nil
}

// unmanagedSettings are what unmanaging an app removes, marked deleted: its
// app-managed setting and the jobs it made. An app no engine looks after has
// nothing to unmanage.
func unmanagedSettings(app *entity.App, now time.Time) ([]*entity.Setting, error) {
	var managed *entity.Setting
	for _, setting := range app.Settings {
		if setting.Type == base.SettingTypeAppManaged && setting.DeletedAt.IsZero() {
			managed = setting
		}
	}
	if managed == nil {
		return nil, hperrors.NewNotFound("Managed engine")
	}
	deleting := []*entity.Setting{managed}
	for _, setting := range app.Settings {
		if setting.Type != base.SettingTypeSchedJob {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if job.ManagedBy.ID == managed.ID {
			deleting = append(deleting, setting)
		}
	}
	for _, setting := range deleting {
		setting.DeletedAt = now
		setting.UpdateVer++
	}
	return deleting, nil
}
