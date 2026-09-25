package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// prepareUpdatedServices keeps, while the transaction is open, the deployment
// each updated app with a service is to be built from after it commits: naming
// a volume by its id reads what the transaction wrote.
func (w *writer) prepareUpdatedServices(ctx context.Context) error {
	for _, node := range w.p.writingApps() {
		app := w.p.targetApps[node.Path]
		if node.Action != specmodel.ActionUpdate || app == nil || app.ServiceID == "" ||
			!changesDeploymentBlocks(node.Changes) {
			continue
		}
		deployment, err := w.preparedDeployment(ctx, node)
		if err != nil {
			return err
		}
		w.prepared[node.Path] = deployment
	}
	return nil
}

// changesDeploymentBlocks reports whether changes reach an app's service spec:
// any deployment block but the source, which only a deployment applies.
func changesDeploymentBlocks(changes []string) bool {
	return slices.ContainsFunc(changes, func(change string) bool {
		return strings.HasPrefix(change, "deployment.") && change != changeDeploymentSource
	})
}

// applyToServices is phase 2: each updated app's running service is brought to
// what was committed - its deployment blocks in one update, its secrets and
// config files, its routing - and then the environment of every app in each env
// touched. A running service cannot be rolled back, which is why this comes
// after the commit. An app that fails is marked so, with its error, and the rest
// go on: its configuration is saved, and importing again retries it.
func (w *writer) applyToServices(ctx context.Context, db database.IDB) {
	envs := map[string]*entity.ObjectScope{}
	byAppID := map[string]*specmodel.PlanNode{}
	for _, node := range w.p.writingApps() {
		app := w.p.targetApps[node.Path]
		if node.Action != specmodel.ActionUpdate || app == nil {
			continue
		}
		byAppID[app.ID] = node
		envs[app.ProjectEnvID] = entity.NewObjectScopeProjectEnv(app.ProjectID, app.ProjectEnvID)
		if app.ServiceID == "" {
			continue
		}
		if err := w.applyToService(ctx, db, node, app); err != nil {
			failNode(node, err)
		}
	}
	if err := w.addEnvsOfChangedVariables(ctx, db, envs); err != nil {
		for _, node := range byAppID {
			failNode(node, err)
		}
		return
	}
	for _, envID := range slices.Sorted(maps.Keys(envs)) {
		data, err := w.p.s.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, envs[envID], false, nil, true, true)
		if err != nil {
			w.failEnv(envID, byAppID, err)
			continue
		}
		for i, applyErr := range w.p.s.envVarService.ApplyEnvVarsForApps(ctx, db, data, true, true) {
			if node := byAppID[data[i].App.ID]; node != nil {
				failNode(node, applyErr)
			}
		}
	}
}

func failNode(node *specmodel.PlanNode, err error) {
	if node.Outcome == specmodel.OutcomeFailed {
		return
	}
	node.Outcome, node.Error = specmodel.OutcomeFailed, err.Error()
}

func (w *writer) failEnv(envID string, byAppID map[string]*specmodel.PlanNode, err error) {
	for id, node := range byAppID {
		if app := w.p.targetApps[node.Path]; app != nil && app.ProjectEnvID == envID {
			failNode(byAppID[id], err)
		}
	}
}

// addEnvsOfChangedVariables adds the envs whose apps read variables or secrets
// an env's or a project's settings changed: the env's own apps, every env's of
// the project.
func (w *writer) addEnvsOfChangedVariables(
	ctx context.Context, db database.IDB, envs map[string]*entity.ObjectScope,
) error {
	for _, node := range w.writtenNodes() {
		if node.Kind != specmodel.NodeKindSettings || !changesVariables(w.writtenNames(node)) {
			continue
		}
		scope, objectID := w.scopeOf(node)
		if scope == base.ObjectScopeProjectEnv {
			projectID, _ := projecthelper.ParseProjectEnvID(objectID)
			envs[objectID] = entity.NewObjectScopeProjectEnv(projectID, objectID)
			continue
		}
		projectEnvs, _, err := w.p.s.projectEnvRepo.List(ctx, db, objectID, nil)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, env := range projectEnvs {
			envs[env.ID] = entity.NewObjectScopeProjectEnv(env.ProjectID, env.ID)
		}
	}
	return nil
}

func changesVariables(names []string) bool {
	return slices.ContainsFunc(names, func(name string) bool {
		block, _, _ := strings.Cut(name, "/")
		return block == specmodel.SingletonBlockName(base.SettingTypeEnvVar) ||
			block == specmodel.CollectionBlockName(base.SettingTypeSecret)
	})
}

// applyToService brings one app's service to its committed configuration.
func (w *writer) applyToService(ctx context.Context, db database.IDB, node *specmodel.PlanNode, app *entity.App) error {
	app, err := w.p.s.appRepo.GetByID(ctx, db, app.ProjectID, app.ID,
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment := w.prepared[node.Path]
	if deployment != nil || slices.Contains(node.Changes, string(specmodel.BlockSettingsDockerAPI)) {
		if err = w.updateService(ctx, db, node, app, deployment); err != nil {
			return err
		}
	}
	if slices.Contains(node.Changes, "settings."+specmodel.SingletonBlockName(base.SettingTypeAppRouting)) {
		routing := app.GetSettingByType(base.SettingTypeAppRouting)
		if routing == nil {
			return nil
		}
		settings, err := routing.AsAppRoutingSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}
		_, err = w.p.s.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
			App: app, RoutingSettings: settings,
		})
		return hperrors.Wrap(err)
	}
	return nil
}

// updateService builds an app's deployment blocks, when they changed, onto its
// running spec, and what its Docker API access needs, and updates the service
// once: one restart, where the settings screens make one update per block.
// Rebuilding storage or networks writes them afresh, without the socket and the
// network the access gives, which is why the access is applied after them.
func (w *writer) updateService(
	ctx context.Context, db database.IDB, node *specmodel.PlanNode, app *entity.App, deployment *specmodel.Deployment,
) error {
	svc, err := w.p.s.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if deployment != nil {
		_, err = w.p.s.BuildApp(ctx, db, &specservice.BuildAppReq{
			App: app, Doc: &specmodel.AppDoc{App: node.Key, Deployment: deployment}, Spec: &svc.Spec,
			TimeNow: w.now, Import: true,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	if err = w.p.s.dockerAPIService.ApplyToService(ctx, db, app.ID, &svc.Spec); err != nil {
		return hperrors.Wrap(err)
	}
	_, err = w.p.s.clusterService.ServiceUpdate(ctx, app.ServiceID, &svc.Version, &svc.Spec)
	return hperrors.Wrap(err)
}
