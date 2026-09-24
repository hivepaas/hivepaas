package specserviceimpl

import (
	"context"
	"slices"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// writeUpdatedApps does, in the transaction, what an existing app's settings
// take beyond their rows: its scheduled jobs that changed are scheduled again,
// and its deployment queued when its source changed and the options ask. Its
// running service is phase 2's.
func (w *writer) writeUpdatedApps(ctx context.Context) error {
	for _, node := range w.p.writingApps() {
		if node.Action != specmodel.ActionUpdate {
			continue
		}
		appID := w.appIDs[node.Path]
		var jobs []*entity.Setting
		var source *entity.Setting
		for _, setting := range w.written {
			switch {
			case setting.ObjectID != appID:
			case setting.Type == base.SettingTypeSchedJob:
				jobs = append(jobs, setting)
			case setting.Type == base.SettingTypeAppDeployment:
				source = setting
			}
		}
		if len(jobs) > 0 {
			tx, _ := w.p.db.(database.Tx)
			if err := w.p.s.taskQueue.ScheduleTasksForSchedJobs(ctx, tx, jobs, true); err != nil {
				return hperrors.Wrap(err)
			}
		}
		if node.Deploy && source != nil {
			if err := w.queueDeployment(ctx, w.p.targetApps[node.Path], source); err != nil {
				return err
			}
		}
	}
	return nil
}

// queueDeployment creates the deployment of an app's source, to be scheduled
// once the transaction has committed.
func (w *writer) queueDeployment(ctx context.Context, app *entity.App, source *entity.Setting) error {
	settings, err := source.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment, task, err := w.p.s.appDeploymentService.CreateDeploymentAndTask(app, settings,
		appdeploymentservice.DeploymentArgs{})
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Trigger = &entity.AppDeploymentTrigger{Source: base.DeploymentTriggerSourceAPI, SourceID: w.operatorID}
	err = w.p.s.appService.PersistAppData(ctx, w.p.db, &appservice.PersistingAppData{
		UpsertingDeployments: []*entity.Deployment{deployment},
		UpsertingTasks:       []*entity.Task{task},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	w.tasks = append(w.tasks, task)
	w.deployments = append(w.deployments, &specservice.ImportDeployment{AppID: app.ID, DeploymentID: deployment.ID})
	return nil
}

// provisionApps creates the new apps, the way template creation creates one:
// each with its service, built from the bundle, its settings, and - when the
// options ask - its first deployment. It runs once the scopes are persisted,
// since an app is created in an env that has to exist.
func (w *writer) provisionApps(ctx context.Context) error {
	var reqs []*appprovisionservice.ProvisionAppReq
	for _, node := range w.createdApps() {
		place := w.p.apps[node.Path]
		deployment, err := w.preparedDeployment(ctx, node)
		if err != nil {
			return err
		}
		projectID := w.projectID(place.project)
		req := &appprovisionservice.ProvisionAppReq{
			ProjectID:    projectID,
			ProjectEnvID: projecthelper.CalcProjectEnvID(projectID, place.env),
			AppID:        w.appIDs[node.Path],
			Key:          node.Key,
			Name:         gofn.Coalesce(place.doc.Name, node.Key),
			Status:       gofn.Coalesce(base.AppStatus(place.doc.Status), base.AppStatusActive),
			Note:         place.doc.Note,
			Configure:    w.configureApp(node, deployment),
		}
		if node.Deploy {
			req.Deployment = &appprovisionservice.FirstDeployment{
				Source: base.DeploymentTriggerSourceAPI, SourceID: w.operatorID,
			}
		}
		reqs = append(reqs, req)
	}
	if len(reqs) == 0 {
		return nil
	}
	provisioned, err := w.p.s.appProvisionService.ProvisionApps(ctx, w.p.db,
		&appprovisionservice.ProvisionAppsReq{Apps: reqs})
	w.provisioned = provisioned
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, one := range provisioned.Apps {
		if one.DeploymentTask != nil {
			w.tasks = append(w.tasks, one.DeploymentTask)
			w.deployments = append(w.deployments, &specservice.ImportDeployment{
				AppID: one.App.ID, DeploymentID: one.Deployment.ID,
			})
		}
		w.tasks = append(w.tasks, one.CertTasks...)
	}
	return nil
}

// createdApps are the apps the import creates, an app whose directory another
// one's mount reaches before that one: building a mount into another app's
// directory finds that app.
func (w *writer) createdApps() []*specmodel.PlanNode {
	var created []*specmodel.PlanNode
	reached := map[string]bool{}
	for _, node := range w.p.writingApps() {
		if node.Action != specmodel.ActionCreate {
			continue
		}
		created = append(created, node)
		if d := w.p.apps[node.Path].doc.Deployment; d != nil && d.Storage != nil {
			for _, m := range d.Storage.Mounts {
				if m.SourceApp != nil {
					reached[m.SourceApp.App] = true
				}
			}
		}
	}
	slices.SortStableFunc(created, func(a, b *specmodel.PlanNode) int {
		switch {
		case reached[a.Key] && !reached[b.Key]:
			return -1
		case reached[b.Key] && !reached[a.Key]:
			return 1
		}
		return 0
	})
	return created
}

// configureApp fills in a new app: its service built from the prepared
// deployment, and the settings the writer built for it.
func (w *writer) configureApp(
	node *specmodel.PlanNode, deployment *specmodel.Deployment,
) appprovisionservice.ConfigureFunc {
	return func(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
		[]*entity.Setting, error) {
		_, err := w.p.s.BuildApp(ctx, db, &specservice.BuildAppReq{
			App:     app,
			Doc:     &specmodel.AppDoc{App: node.Key, Deployment: deployment},
			Spec:    spec,
			TimeNow: w.now,
			Import:  true,
		})
		if err != nil {
			return nil, hperrors.Wrap(err).WithExtraDetail("%s", node.Path)
		}
		return w.appSettings[node.Path], nil
	}
}

// cleanup removes from docker what provisioning created, for a transaction
// that did not commit.
func (w *writer) cleanup(ctx context.Context) error {
	if w.provisioned == nil || w.provisioned.Cleanup == nil {
		return nil
	}
	return hperrors.Wrap(w.provisioned.Cleanup(ctx))
}
