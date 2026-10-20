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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedscripts"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// UpgradeManagedScripts replaces a managed database's scripts and artifacts with
// what its template's revision offers, on the database, its members and its
// verifier, in one transaction; then deploys each, the member that writes last
// (docs/superpowers/specs/2026-10-21-managed-scripts-upgrade-design.md §2).
func (uc *UC) UpgradeManagedScripts(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UpgradeManagedScriptsReq,
) (*appmanageddto.UpgradeManagedScriptsResp, error) {
	data := &appmanageddto.UpgradeManagedScriptsDataResp{Deployments: []*appmanageddto.ManagedDeploymentResp{}}
	var tasks []*entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		tasks, txErr = uc.upgradeScripts(ctx, db, auth, req, data)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, tasks...); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.UpgradeManagedScriptsResp{Data: data}, nil
}

func (uc *UC) upgradeScripts(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appmanageddto.UpgradeManagedScriptsReq,
	data *appmanageddto.UpgradeManagedScriptsDataResp,
) ([]*entity.Task, error) {
	loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return nil, err
	}
	if loaded.managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", loaded.app.Name).ExposeParams("App")
	}
	if state := loaded.managed.Database; state != nil && state.Restore != nil && state.Restore.ReadyAt.IsZero() {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", loaded.app.Name).ExposeParams("App")
	}
	view, err := uc.scriptsOf(ctx, db, loaded)
	if err != nil {
		return nil, err
	}
	if err = checkUpgradable(view, req.Fingerprint); err != nil {
		return nil, err
	}
	now := timeutil.NowUTC()
	record := &entity.ManagedScripts{Revision: view.AvailableRevision,
		Fingerprint: view.Offer.Available.Fingerprint(), UpdatedAt: now}
	apps, err := uc.clusterApps(ctx, db, loaded)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		if err = uc.upgradeAppScripts(ctx, db, loaded, app, view.Offer.Available, record, now); err != nil {
			return nil, err
		}
	}
	data.Revision, data.Fingerprint = record.Revision, record.Fingerprint
	err = auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type: base.AuditLogTypeAppUpdate, Scope: base.ObjectScopeApp, ObjectID: loaded.app.ID,
		Source: base.AuditLogSourceAPIUpdate, Section: auditSectionManaged, Auth: auth,
		ResType: base.ResourceTypeApp, ResID: loaded.app.ID, ResName: loaded.app.Name,
		Detail: auditdetail.New().Set("scriptsUpgraded", record.Revision).Set("fingerprint", record.Fingerprint).
			Set("changes", len(view.Offer.Changes)).Set("apps", len(apps)).String(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	writer := loaded.app.ID
	if state := loaded.managed.Database; state != nil && state.Writer.ID != "" {
		writer = state.Writer.ID
	}
	return uc.deployScripts(ctx, db, auth, deployOrder(apps, writer), data)
}

// checkUpgradable refuses an upgrade with nothing to apply - unless a member is
// out of step, which applying again brings back - and one whose reviewed
// fingerprint the revision no longer offers.
func checkUpgradable(view *appmanageddto.ManagedScriptsView, fingerprint string) error {
	offer := view.Offer
	switch {
	case offer.State == managedscripts.StateUpdateAvailable:
	case offer.State == managedscripts.StateUpToDate && len(view.OutOfStep) > 0:
	default:
		return hperrors.NewArgumentInvalid("fingerprint").
			WithExtraDetail("the scripts are %s, with nothing to apply %s", offer.State, offer.Reason)
	}
	if fingerprint != offer.Available.Fingerprint() {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched).
			WithExtraDetail("the templates offer other scripts than those reviewed: review them again")
	}
	return nil
}

// clusterApps are the cluster's app, then its members and verifier.
func (uc *UC) clusterApps(ctx context.Context, db database.IDB, loaded *loadedManaged) ([]*entity.App, error) {
	apps := []*entity.App{loaded.app}
	for _, member := range clusterMembers(loaded.managed) {
		app, err := uc.appService.LoadApp(ctx, db, loaded.app.ProjectID, member.App.ID, false, false,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// upgradeAppScripts replaces one app's scripts and artifacts, and records the
// upgrade on its engine setting with what its locks gain and lose. A verifier
// has no engine setting: its scripts change alone.
func (uc *UC) upgradeAppScripts(ctx context.Context, db database.Tx, loaded *loadedManaged, app *entity.App,
	available *managedscripts.Set, record *entity.ManagedScripts, now time.Time) error {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectWhereIn("setting.type IN (?)", scriptSettingTypes...),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	patch, err := managedscripts.Plan(app, settings, available, now)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range patch.Created {
		if err = uc.settingRepo.Insert(ctx, db, setting); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range patch.Updated {
		if err = uc.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range patch.Deleted {
		if err = uc.settingRepo.Update(ctx, db, setting, bunex.UpdateColumns("deleted_at", "update_ver")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	setting, managed := loaded.setting, loaded.managed
	if app.ID != loaded.app.ID {
		if setting, managed, err = uc.managedSettingOf(ctx, db, app.ID); err != nil || setting == nil {
			return err
		}
	}
	managed.Locks.Settings = slices.DeleteFunc(managed.Locks.Settings, func(id string) bool {
		return slices.Contains(patch.LocksRemoved, id)
	})
	managed.Locks.Settings = append(managed.Locks.Settings, patch.LocksAdded...)
	recorded := *record
	managed.Scripts = &recorded
	return uc.saveManagedSetting(ctx, db, setting, managed, now)
}

// managedSettingOf is an app's engine setting, nil for an app without one.
func (uc *UC) managedSettingOf(ctx context.Context, db database.IDB, appID string) (
	*entity.Setting, *entity.AppManagedSettings, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil || len(settings) == 0 {
		return nil, nil, hperrors.Wrap(err)
	}
	managed, err := settings[0].AsAppManagedSettings()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return settings[0], managed, nil
}

// deployOrder is the cluster's apps with the member that writes last: the
// others restart onto their new scripts while it still serves.
func deployOrder(apps []*entity.App, writer string) []*entity.App {
	out := make([]*entity.App, 0, len(apps))
	var last *entity.App
	for _, app := range apps {
		if app.ID == writer {
			last = app
			continue
		}
		out = append(out, app)
	}
	if last != nil {
		out = append(out, last)
	}
	return out
}

// deployScripts makes a deployment of each app, in order, and returns the tasks
// to queue. A stopped service - a fenced member, a verifier between runs - is
// deployed too: its service takes the new files at no replicas, and starts with
// them.
func (uc *UC) deployScripts(ctx context.Context, db database.Tx, auth *basedto.Auth, apps []*entity.App,
	data *appmanageddto.UpgradeManagedScriptsDataResp) ([]*entity.Task, error) {
	var tasks []*entity.Task
	persisting := &appservice.PersistingAppData{}
	for _, app := range apps {
		setting, err := uc.settingRepo.GetSingle(ctx, db, app.GetObjectScope(), base.SettingTypeAppDeployment, true,
			bunex.SelectWhere("setting.object_id = ?", app.ID))
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		settings, err := setting.AsAppDeploymentSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, settings,
			appdeploymentservice.DeploymentArgs{})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		deployment.Trigger = &entity.AppDeploymentTrigger{Source: base.DeploymentTriggerSourceAPI,
			SourceID: auth.User.ID}
		persisting.UpsertingDeployments = append(persisting.UpsertingDeployments, deployment)
		persisting.UpsertingTasks = append(persisting.UpsertingTasks, task)
		tasks = append(tasks, task)
		data.Deployments = append(data.Deployments, &appmanageddto.ManagedDeploymentResp{
			App:        &basedto.ObjectIDResp{ID: app.ID},
			Deployment: &basedto.ObjectIDResp{ID: deployment.ID},
			Task:       &basedto.ObjectIDResp{ID: task.ID},
		})
	}
	if err := uc.appService.PersistAppData(ctx, db, persisting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return tasks, nil
}
