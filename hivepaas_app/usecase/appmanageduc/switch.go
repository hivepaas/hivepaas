package appmanageduc

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/envvarhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// switchConsumer is an app of the environment that names a source database:
// the settings that do, and how.
type switchConsumer struct {
	app *entity.App
	// envSetting and envKeys are the app's variables and those of them that
	// name the source's.
	envSetting *entity.Setting
	envKeys    []string
	// features is the app's features setting, when its previews clone the
	// source.
	features     *entity.Setting
	previewClone bool
	// deployment is the app's deployment setting, when its deployments mark a
	// restore point on the source.
	deployment         *entity.Setting
	deployRestorePoint bool
}

// switchState is a switch's databases and the source's consumers, loaded.
type switchState struct {
	target    *loadedManaged
	source    *entity.App
	consumers []*switchConsumer
}

// GetManagedSwitch says what switching to a restored database would do: which
// database it was restored from, and which apps of the environment use it.
func (uc *UC) GetManagedSwitch(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmanageddto.GetManagedSwitchReq,
) (*appmanageddto.GetManagedSwitchResp, error) {
	state, err := uc.loadSwitch(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	data := &appmanageddto.ManagedSwitchDataResp{
		Source:     switchApp(state.source),
		Consumers:  make([]*appmanageddto.ManagedSwitchConsumerResp, 0, len(state.consumers)),
		SwitchedAt: switchedAt(state.target.managed.Postgres.Restore.SwitchedAt),
	}
	for _, consumer := range state.consumers {
		data.Consumers = append(data.Consumers, &appmanageddto.ManagedSwitchConsumerResp{
			App:                switchApp(consumer.app),
			EnvVars:            append([]string{}, consumer.envKeys...),
			PreviewClone:       consumer.previewClone,
			DeployRestorePoint: consumer.deployRestorePoint,
		})
	}
	return &appmanageddto.GetManagedSwitchResp{Data: data}, nil
}

// SwitchManaged switches the consumers chosen from the database this one was
// restored from to this one. Their variables, their previews' databases and their
// deployments' restore point name this one instead. Their variables are applied
// to their running services, as saving them does - a deployment does not apply
// them - and with redeploy, each whose variables changed is deployed too, for
// what reads the database as it builds or in its pre-deploy command. The source
// keeps running, its jobs paused, so that two databases do not both back up as
// the one in use.
func (uc *UC) SwitchManaged(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.SwitchManagedReq,
) (*appmanageddto.SwitchManagedResp, error) {
	resp := &appmanageddto.SwitchManagedResp{Meta: &basedto.Meta{}, Data: &appmanageddto.SwitchManagedDataResp{
		Switched: []*basedto.ObjectIDResp{}, Deployments: []*appmanageddto.ManagedSwitchDeploymentResp{}}}
	var envData []*envvarservice.AppEnvVarData
	var tasks []*entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		state, err := uc.loadSwitch(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
		if err != nil {
			return err
		}
		chosen, err := chosenConsumers(state.consumers, req.Consumers)
		if err != nil {
			return err
		}
		now := timeutil.NowUTC()
		var changedEnv []*entity.App
		for _, consumer := range chosen {
			for _, setting := range switchConsumerTo(consumer, state.source, state.target.app) {
				setting.UpdatedAt = now
				if err = uc.settingRepo.Update(ctx, db, setting,
					bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
					return hperrors.Wrap(err)
				}
			}
			resp.Data.Switched = append(resp.Data.Switched, &basedto.ObjectIDResp{ID: consumer.app.ID})
			if len(consumer.envKeys) > 0 {
				changedEnv = append(changedEnv, consumer.app)
			}
		}
		if err = uc.pauseSource(ctx, db, state, now); err != nil {
			return err
		}
		if req.Redeploy {
			if tasks, err = uc.redeployConsumers(ctx, db, auth, changedEnv, resp.Data); err != nil {
				return err
			}
		}
		// The variables are built here, which checks their references resolve,
		// and applied once committed.
		if envData, err = uc.buildConsumersEnv(ctx, db, state.target.app, changedEnv); err != nil {
			return err
		}
		return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: state.target.app.ID,
			Source:   base.AuditLogSourceAPIUpdate,
			Section:  auditSectionManaged,
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    state.target.app.ID,
			ResName:  state.target.app.Name,
			Detail: auditdetail.New().Set("switchedFrom", state.source.ID).
				Set("consumers", len(chosen)).Set("redeploy", req.Redeploy).String(),
		}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The variables first: a deployment queued after updates the service they
	// are on.
	errs := uc.envVarService.ApplyEnvVarsForApps(ctx, uc.db, envData, true, true)
	for i, applyErr := range errs {
		resp.Meta.Warning += fmt.Sprintf("App '%v': %v\n", envData[i].App.Name, applyErr.Error())
	}
	if resp.Meta.Warning != "" {
		resp.Meta.Warning = "The consumers were switched, but not all of them could be applied:\n" + resp.Meta.Warning
	}
	if err = uc.taskQueue.ScheduleTask(ctx, tasks...); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

// loadSwitch loads a restored database, the database it was restored from, and
// the apps of their environment that use the source.
func (uc *UC) loadSwitch(
	ctx context.Context,
	db database.IDB,
	projectID, appID string,
	extraOpts ...bunex.SelectQueryOption,
) (*switchState, error) {
	target, err := uc.loadManaged(ctx, db, projectID, appID, extraOpts...)
	if err != nil {
		return nil, err
	}
	if err = checkSwitchable(target); err != nil {
		return nil, err
	}
	apps, _, err := uc.appRepo.List(ctx, db, projectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectWhere("app.project_env_id = ?", target.app.ProjectEnvID),
		bunex.SelectRelation("Settings",
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeEnvVar, base.SettingTypeAppFeatures,
				base.SettingTypeAppDeployment, base.SettingTypeAppManaged, base.SettingTypeSchedJob),
			bunex.SelectWhereIn("setting.status IN (?)", base.SettingStatusActive, base.SettingStatusDisabled),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	state := &switchState{target: target}
	for _, app := range apps {
		if app.ID == target.managed.Postgres.Restore.SourceApp.ID {
			state.source = app
		}
	}
	if state.source == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchSourceGone).
			WithParam("App", target.app.Name).ExposeParams("App")
	}
	state.consumers = findConsumers(apps, state.source, target.app.ID)
	return state, nil
}

// checkSwitchable refuses a database that was not restored from another, or is
// still being restored.
func checkSwitchable(target *loadedManaged) error {
	pg := target.managed.Postgres
	if pg == nil || pg.Restore == nil || pg.Restore.SourceApp.ID == "" {
		return hperrors.Wrap(hperrors.ErrAppManagedSwitchNotRestored).
			WithParam("App", target.app.Name).ExposeParams("App")
	}
	if pg.Restore.ReadyAt.IsZero() {
		return hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", target.app.Name).ExposeParams("App")
	}
	return nil
}

// findConsumers are the apps of the environment that name the source, by its
// key in a variable that is not literal, by its id among their previews'
// databases to clone, or as their deployments' restore point. The source, the
// restored database and previews, which keep the variables they were made
// with, are not among them.
func findConsumers(apps []*entity.App, source *entity.App, targetID string) []*switchConsumer {
	var out []*switchConsumer
	for _, app := range apps {
		if app.ID == source.ID || app.ID == targetID || app.IsPreviewApp() {
			continue
		}
		consumer := &switchConsumer{app: app}
		for _, setting := range app.Settings {
			if setting.Status != base.SettingStatusActive || !setting.DeletedAt.IsZero() {
				continue
			}
			switch setting.Type { //nolint:exhaustive // three types name another app
			case base.SettingTypeEnvVar:
				consumer.readEnv(setting, source.Key)
			case base.SettingTypeAppFeatures:
				features, err := setting.AsAppFeatureSettings()
				if err == nil && features.PreviewSettings != nil &&
					slices.Contains(features.PreviewSettings.AppsToClone.ToIDStringSlice(), source.ID) {
					consumer.features, consumer.previewClone = setting, true
				}
			case base.SettingTypeAppDeployment:
				deployment, err := setting.AsAppDeploymentSettings()
				if err == nil && deployment.RestorePointBeforeDeploy.ID == source.ID {
					consumer.deployment, consumer.deployRestorePoint = setting, true
				}
			}
		}
		if len(consumer.envKeys) > 0 || consumer.previewClone || consumer.deployRestorePoint {
			out = append(out, consumer)
		}
	}
	return out
}

func (c *switchConsumer) readEnv(setting *entity.Setting, sourceKey string) {
	envVars, err := setting.AsEnvVars()
	if err != nil {
		return
	}
	for _, env := range envVars.Data {
		if envvarhelper.RefersTo(env, sourceKey) {
			c.envKeys = append(c.envKeys, env.Key)
		}
	}
	if len(c.envKeys) > 0 {
		c.envSetting = setting
	}
}

// chosenConsumers are the consumers a request names: each must be one.
func chosenConsumers(consumers []*switchConsumer, chosen []*basedto.ObjectIDReq) ([]*switchConsumer, error) {
	byID := make(map[string]*switchConsumer, len(consumers))
	for _, consumer := range consumers {
		byID[consumer.app.ID] = consumer
	}
	out := make([]*switchConsumer, 0, len(chosen))
	seen := map[string]bool{}
	for _, ref := range chosen {
		consumer := byID[ref.ID]
		if consumer == nil {
			return nil, hperrors.NewArgumentInvalid("consumers").
				WithExtraDetail("%s does not use the database this one was restored from", ref.ID)
		}
		if !seen[ref.ID] {
			seen[ref.ID] = true
			out = append(out, consumer)
		}
	}
	return out, nil
}

// switchConsumerTo repoints a consumer's settings from the source to the
// target, and returns those it changed, each a new version.
func switchConsumerTo(consumer *switchConsumer, source, target *entity.App) []*entity.Setting {
	var changed []*entity.Setting
	if consumer.envSetting != nil {
		envVars := consumer.envSetting.MustAsEnvVars()
		replacer := envvarhelper.RefReplacer(map[string]string{source.Key: target.Key})
		if len(envvarhelper.RewriteRefs(envVars.Data, replacer)) > 0 {
			consumer.envSetting.MustSetData(envVars)
			changed = append(changed, consumer.envSetting)
		}
	}
	if consumer.features != nil {
		features := consumer.features.MustAsAppFeatureSettings()
		for _, ref := range features.PreviewSettings.AppsToClone {
			if ref.ID == source.ID {
				ref.ID = target.ID
			}
		}
		consumer.features.MustSetData(features)
		changed = append(changed, consumer.features)
	}
	if consumer.deployment != nil {
		deployment := consumer.deployment.MustAsAppDeploymentSettings()
		deployment.RestorePointBeforeDeploy = entity.ObjectID{ID: target.ID}
		consumer.deployment.MustSetData(deployment)
		changed = append(changed, consumer.deployment)
	}
	for _, setting := range changed {
		setting.UpdateVer++
	}
	return changed
}

// pauseSource pauses the source's managed jobs and records the switch on both
// databases.
func (uc *UC) pauseSource(ctx context.Context, db database.Tx, state *switchState, now time.Time) error {
	for _, setting := range state.source.Settings {
		if setting.Type != base.SettingTypeAppManaged || !setting.DeletedAt.IsZero() {
			continue
		}
		managed, err := setting.AsAppManagedSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, job := range pauseManagedJobs(state.source.Settings, setting.ID, now) {
			if err = uc.settingRepo.Update(ctx, db, job,
				bunex.UpdateColumns("status", "update_ver", "updated_at")); err != nil {
				return hperrors.Wrap(err)
			}
			if err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, job, true); err != nil {
				return hperrors.Wrap(err)
			}
		}
		managed.SwitchedTo = &entity.ManagedSwitch{App: entity.ObjectID{ID: state.target.app.ID}, At: now}
		if err = uc.saveManagedSetting(ctx, db, setting, managed, now); err != nil {
			return err
		}
	}
	state.target.managed.Postgres.Restore.SwitchedAt = now
	return uc.saveManagedSetting(ctx, db, state.target.setting, state.target.managed, now)
}

func (uc *UC) saveManagedSetting(ctx context.Context, db database.Tx, setting *entity.Setting,
	managed *entity.AppManagedSettings, now time.Time) error {
	if err := setting.SetData(managed); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = now
	return hperrors.Wrap(uc.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("data", "update_ver", "updated_at")))
}

// pauseManagedJobs disables the active jobs the engine's setting made, and
// returns them.
func pauseManagedJobs(settings []*entity.Setting, managedID string, now time.Time) []*entity.Setting {
	var out []*entity.Setting
	for _, setting := range settings {
		if setting.Type != base.SettingTypeSchedJob || setting.Status != base.SettingStatusActive ||
			!setting.DeletedAt.IsZero() {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil || job.ManagedBy.ID != managedID {
			continue
		}
		setting.Status = base.SettingStatusDisabled
		setting.UpdateVer++
		setting.UpdatedAt = now
		out = append(out, setting)
	}
	return out
}

// redeployConsumers makes a deployment of each app that has a deployment
// setting to deploy by, and returns the tasks to queue.
func (uc *UC) redeployConsumers(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	apps []*entity.App,
	resp *appmanageddto.SwitchManagedDataResp,
) ([]*entity.Task, error) {
	var tasks []*entity.Task
	persisting := &appservice.PersistingAppData{}
	for _, app := range apps {
		setting := app.GetSettingByType(base.SettingTypeAppDeployment)
		if setting == nil || !setting.IsActive() {
			continue
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
		resp.Deployments = append(resp.Deployments, &appmanageddto.ManagedSwitchDeploymentResp{
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

// buildConsumersEnv builds the apps' variables as stored now, which checks that
// what they name resolves, for them to be applied once committed.
func (uc *UC) buildConsumersEnv(
	ctx context.Context,
	db database.Tx,
	target *entity.App,
	apps []*entity.App,
) ([]*envvarservice.AppEnvVarData, error) {
	if len(apps) == 0 {
		return nil, nil
	}
	projectEnv := *target.ProjectEnv
	projectEnv.Project = target.Project
	projectEnv.Apps = apps
	ids := make([]string, 0, len(apps))
	for _, app := range apps {
		ids = append(ids, app.ID)
	}
	data, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, projectEnv.GetObjectScope(), false, ids,
		false, false)
	return data, hperrors.Wrap(err)
}

func switchApp(app *entity.App) *appmanageddto.ManagedSwitchAppResp {
	return &appmanageddto.ManagedSwitchAppResp{ID: app.ID, Name: app.Name, Key: app.Key}
}

func switchedAt(at time.Time) *time.Time {
	if at.IsZero() {
		return nil
	}
	return &at
}
