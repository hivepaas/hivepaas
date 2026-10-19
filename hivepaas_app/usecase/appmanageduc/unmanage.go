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
// they are the person's from now on. A verifier app the restore test ran in goes
// with it: it has nothing left to verify.
func (uc *UC) UnmanageApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UnmanageAppReq,
) (*appmanageddto.UnmanageAppResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectFor("UPDATE OF app"),
			// What LoadApp checks the project and its environment are active by.
			bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
			bunex.SelectRelation("ProjectEnv"),
			bunex.SelectRelation("Settings",
				bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppManaged, base.SettingTypeSchedJob),
				// A restored database's jobs are disabled until it has recovered.
				bunex.SelectWhereIn("setting.status IN (?)", base.SettingStatusActive, base.SettingStatusDisabled),
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
		// Last: a deletion reaches past the database.
		if err = uc.deleteVerifier(ctx, db, app, verifierOf(deleting[0])); err != nil {
			return err
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
	// A member is removed from its primary; a primary loses its members first.
	data, err := managed.AsAppManagedSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if data.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).WithParam("App", app.Name).ExposeParams("App")
	}
	if data.Database != nil && len(data.Database.Members) > 0 {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedHasMembers).WithParam("App", app.Name).ExposeParams("App")
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

// verifierOf is the verifier app an app-managed setting has its restore test run
// in, empty for none.
func verifierOf(managed *entity.Setting) string {
	settings, err := managed.AsAppManagedSettings()
	if err != nil || settings.Database == nil {
		return ""
	}
	return settings.Database.RestoreTest.VerifierApp.ID
}
