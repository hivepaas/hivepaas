package specserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// The fakes apply reaches, beside the export fixture's.

type fakeAppService struct {
	appservice.Service
	persisted []*appservice.PersistingAppData
}

func (f *fakeAppService) PersistAppData(_ context.Context, _ database.IDB, data *appservice.PersistingAppData) error {
	f.persisted = append(f.persisted, data)
	return nil
}

type fakeDeploymentService struct {
	appdeploymentservice.Service
}

func (f *fakeDeploymentService) CreateDeploymentAndTask(
	app *entity.App, settings *entity.AppDeploymentSettings, _ appdeploymentservice.DeploymentArgs,
) (*entity.Deployment, *entity.Task, error) {
	return &entity.Deployment{ID: "dep_" + app.ID, AppID: app.ID, Settings: settings},
		&entity.Task{ID: "task_" + app.ID, ObjectID: app.ID}, nil
}

// fakeProvisionService provisions the way the real one does as far as import
// can tell: it configures each app on an empty spec, and records what it made.
type fakeProvisionService struct {
	appprovisionservice.Service
	reqs      []*appprovisionservice.ProvisionAppReq
	settings  map[string][]*entity.Setting
	specs     map[string]*swarm.ServiceSpec
	cleanedUp bool
}

func (f *fakeProvisionService) ProvisionApps(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppsReq,
) (*appprovisionservice.ProvisionAppsResp, error) {
	resp := &appprovisionservice.ProvisionAppsResp{Cleanup: func(context.Context) error {
		f.cleanedUp = true
		return nil
	}}
	f.settings, f.specs = map[string][]*entity.Setting{}, map[string]*swarm.ServiceSpec{}
	for _, one := range req.Apps {
		f.reqs = append(f.reqs, one)
		app := &entity.App{
			ID: one.AppID, Key: one.Key, Name: one.Name, ProjectID: one.ProjectID, ProjectEnvID: one.ProjectEnvID,
			Project:    &entity.Project{ID: one.ProjectID},
			ProjectEnv: &entity.ProjectEnv{ID: one.ProjectEnvID},
		}
		spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}}
		settings, err := one.Configure(ctx, db, app, spec)
		if err != nil {
			return resp, err
		}
		f.settings[app.ID], f.specs[app.ID] = settings, spec
		provisioned := &appprovisionservice.ProvisionAppResp{App: app}
		if one.Deployment != nil {
			provisioned.Deployment = &entity.Deployment{ID: "dep_" + app.ID, AppID: app.ID}
			provisioned.DeploymentTask = &entity.Task{ID: "task_" + app.ID}
		}
		resp.Apps = append(resp.Apps, provisioned)
	}
	return resp, nil
}

type fakeTaskQueue struct {
	queue.TaskQueue
	scheduledJobs []*entity.Setting
}

func (f *fakeTaskQueue) ScheduleTasksForSchedJobs(
	_ context.Context, _ database.Tx, jobs []*entity.Setting, _ bool,
) error {
	f.scheduledJobs = append(f.scheduledJobs, jobs...)
	return nil
}

// BuildAppMounts mounts each volume by its id, which is what import names it by.
func (f *fakeExportVolumeService) BuildAppMounts(
	_ context.Context, _ database.IDB, req *volumeservice.BuildAppMountsReq,
) (*volumeservice.BuildAppMountsResp, error) {
	mounts := append([]mount.Mount{}, req.Kept...)
	for _, one := range req.New {
		mounts = append(mounts, mount.Mount{Type: one.Type, Source: one.Source, Target: one.Target})
	}
	return &volumeservice.BuildAppMountsResp{Mounts: mounts}, nil
}
