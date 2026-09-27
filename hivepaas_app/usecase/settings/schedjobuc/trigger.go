package schedjobuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// checkTriggerApps refuses triggers naming the wrong apps: an app's job listens
// to its own app and names none; an env job names apps of its env.
func (uc *UC) checkTriggerApps(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	job *entity.SchedJob,
) error {
	if len(job.Triggers) == 0 {
		return nil
	}
	apps := map[string]*entity.App{}
	if appIDs := job.TriggerAppIDs(); len(appIDs) > 0 {
		loaded, err := uc.appService.LoadApps(ctx, db, "", appIDs, false, false)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, app := range loaded {
			apps[app.ID] = app
		}
	}
	for i, trigger := range job.Triggers {
		if problem := triggerAppProblem(scope, trigger, apps); problem != "" {
			return hperrors.NewArgumentInvalid(fmt.Sprintf("triggers[%d].apps", i)).
				WithExtraDetail("trigger %d: %s", i+1, problem)
		}
	}
	return nil
}

// triggerAppProblem says why a trigger's apps do not fit the job's scope, or
// nothing when they do. apps are the apps found, by ID.
func triggerAppProblem(
	scope *entity.ObjectScope,
	trigger *entity.SchedJobTrigger,
	apps map[string]*entity.App,
) string {
	if scope.ScopeType != base.ObjectScopeProjectEnv {
		if len(trigger.Apps) > 0 {
			return "an app's job listens to its own app: name no app"
		}
		return ""
	}
	if len(trigger.Apps) == 0 {
		return "name the apps whose events run the job"
	}
	for _, ref := range trigger.Apps {
		app := apps[ref.ID]
		switch {
		case app == nil:
			return fmt.Sprintf("app %s is not found", ref.ID)
		case app.ProjectEnvID != scope.ProjectEnvID:
			return fmt.Sprintf("app %s is not in this env", ref.ID)
		}
	}
	return ""
}

// loadTriggerApps adds to refObjects the apps the triggers of jobSettings name,
// whatever their scope, for their names in the response.
func (uc *UC) loadTriggerApps(
	ctx context.Context,
	db database.IDB,
	refObjects *entity.RefObjects,
	jobSettings ...*entity.Setting,
) error {
	var appIDs []string
	for _, setting := range jobSettings {
		job, err := setting.AsSchedJob()
		if err != nil {
			return hperrors.Wrap(err)
		}
		appIDs = append(appIDs, job.TriggerAppIDs()...)
	}
	if len(appIDs) == 0 || refObjects == nil {
		return nil
	}
	err := uc.SettingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, false,
		&entity.RefObjectIDs{RefAppIDs: appIDs})
	return hperrors.Wrap(err)
}
