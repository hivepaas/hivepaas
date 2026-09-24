package appprovisionserviceimpl

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

const (
	dockerImageInit    = "busybox:latest"
	dockerImageInitDev = "crccheck/hello-world:latest"
)

func (s *service) ProvisionApp(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
) (resp *appprovisionservice.ProvisionAppResp, err error) {
	project, projectEnv, err := s.loadProjectEnv(ctx, db, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// From here on the response is returned even when provisioning fails: it
	// carries what was created in docker, which no rollback of the caller's
	// transaction can undo. Created stays nil until there is something to undo.
	resp = &appprovisionservice.ProvisionAppResp{}

	timeNow := timeutil.NowUTC()
	app, err := s.newApp(ctx, db, req, project, projectEnv, timeNow)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}

	svc := initialService(app, s.networkService.GetProjectNetworkName(project, projectEnv.Name))
	settings := defaultAppSettings(app, timeNow)
	if req.Configure != nil {
		configured, configureErr := req.Configure(ctx, db, app, &svc.Spec)
		if configureErr != nil {
			return resp, hperrors.Wrap(configureErr)
		}
		settings = replaceSettingsByType(settings, configured)
	}
	app.Settings = settings

	// After Configure, so volume pins are resolved from the mounts it wrote.
	_, err = s.placementService.ApplyPlacementSettings(ctx, db, &placementservice.ApplyPlacementSettingsReq{
		App:                app,
		Service:            svc,
		SkipSavingToDocker: true,
	})
	if err != nil {
		return resp, hperrors.Wrap(err)
	}

	createdSvc, err := s.dockerManager.ServiceCreate(ctx, &svc.Spec)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	if createdSvc.ID == "" { // should never happen
		return resp, hperrors.Wrap(hperrors.ErrInfraInternal).WithParam("Error", "empty service ID returned")
	}
	app.ServiceID = createdSvc.ID
	resp.App = app
	resp.Created = &appprovisionservice.CreatedInDocker{ServiceID: createdSvc.ID}

	// What this call made in docker is undone here, where the app it belongs to
	// has no record yet. Once it has one, a failure is the caller's: the records
	// go with its transaction and what is in docker does not, which is what the
	// response carries for it.
	persisted := false
	defer func() {
		if err != nil && !persisted {
			_ = s.clusterService.ServiceRemove(context.WithoutCancel(ctx), app.ServiceID,
				clusterservice.ItemRemovalRetryMax, 0)
			resp.Created = nil
		}
	}()

	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingApps:     []*entity.App{app},
		UpsertingTags:     appTags(app, req.Tags),
		UpsertingSettings: settings,
	})
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	persisted = true

	applied, err := s.ApplyAppConfiguration(ctx, db, &appprovisionservice.ApplyAppConfigurationReq{App: app})
	if applied != nil {
		resp.Created.Configs, resp.Created.Secrets = applied.Configs, applied.Secrets
		resp.CertTasks = applied.CertTasks
	}
	if err != nil {
		return resp, hperrors.Wrap(err)
	}

	if err = s.createFirstDeployment(ctx, db, req, resp); err != nil {
		return resp, hperrors.Wrap(err)
	}
	return resp, nil
}

// createFirstDeployment queues the deployment that replaces the placeholder
// service image with the one the app's deployment settings name.
//
// The task it writes is not scheduled: a task row can be picked up only once the
// transaction it was written in has committed, and that is the caller's.
func (s *service) createFirstDeployment(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
	resp *appprovisionservice.ProvisionAppResp,
) error {
	if req.Deployment == nil {
		return nil
	}
	app := resp.App
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		// The caller asked for a deployment and its Configure wrote no deployment
		// settings, which is a mistake in the caller rather than in the request.
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("app '%s' has no deployment settings to deploy", app.Name)
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	deployment, task, err := s.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings,
		appdeploymentservice.DeploymentArgs{})
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   req.Deployment.Source,
		SourceID: req.Deployment.SourceID,
	}
	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingDeployments: []*entity.Deployment{deployment},
		UpsertingTasks:       []*entity.Task{task},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	resp.Deployment, resp.DeploymentTask = deployment, task
	return nil
}

func (s *service) loadProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
) (*entity.Project, *entity.ProjectEnv, error) {
	project, err := s.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("ProjectEnvs",
			bunex.SelectWhere("project_env.id = ?", req.ProjectEnvID),
		),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if project.Status != base.ProjectStatusActive {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectInactive).WithParam("Name", project.Name)
	}
	if len(project.ProjectEnvs) == 0 {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectEnvNotFound).WithParam("Name", req.ProjectEnvID)
	}
	projectEnv := project.ProjectEnvs[0]
	if projectEnv.Status != base.ProjectStatusActive {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectEnvInactive).WithParam("Project", project.Name).
			WithParam("Env", projectEnv.Name)
	}
	return project, projectEnv, nil
}

func (s *service) newApp(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
	project *entity.Project,
	projectEnv *entity.ProjectEnv,
	timeNow time.Time,
) (*entity.App, error) {
	id := req.AppID
	if id == "" {
		id = gofn.Must(ulid.NewStringULID())
	}
	app := &entity.App{
		ID:              id,
		LogicalParentID: req.LogicalParentID,
		ProjectID:       project.ID,
		Project:         project,
		ProjectEnvID:    projectEnv.ID,
		ProjectEnv:      projectEnv,
		Key:             gofn.Coalesce(req.Key, projecthelper.CalcAppKey(req.Name)),
		Name:            req.Name,
		Status:          req.Status,
		Note:            req.Note,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	app.GlobalKey = projecthelper.CalcAppGlobalKey(project.Key, app.Key, projectEnv.Key)

	// App keys must be unique globally
	conflictApp, err := s.appRepo.GetByGlobalKey(ctx, db, "", app.GlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if conflictApp != nil {
		return nil, hperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", app.GlobalKey)
	}

	// Create local network for the app to attach
	if _, _, err = s.networkService.GetOrCreateProjectNetwork(ctx, db, project, projectEnv.Key); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return app, nil
}

// initialService is the placeholder service an app starts with. Its image is
// replaced by the app's first deployment.
func initialService(app *entity.App, networkName string) *swarm.Service {
	isDevEnv := config.Current().IsDevEnv()
	appInfo := &apphelper.AppInfo{
		Name: app.Name,
		Key:  app.Key,
		Env:  app.ProjectEnv.Name,
	}
	return &swarm.Service{
		Spec: swarm.ServiceSpec{
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: new(uint64(1)),
				},
			},
			Annotations: swarm.Annotations{
				Name: app.GlobalKey,
				Labels: map[string]string{
					appservice.LabelAppNamespace: app.Project.Key,
					appservice.LabelAppInfo:      apphelper.CalcAppInfoLabel(appInfo),
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:    gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
					Command:  gofn.If(isDevEnv, nil, []string{"sleep", "infinity"}),
					Hostname: app.Key,
					// Init is left undecided on purpose: whether a container needs
					// docker's init depends on whether its image starts with one of
					// its own, and the image is not here yet - the first deployment
					// pulls or builds it, and decides then. See applyContainerInit.
					// The app's identity, for the log collector. Container labels
					// rather than service ones: swarm does not pass service labels
					// down, so only these can reach a log line's attrs.
					Labels: appservice.WithAppLogLabels(nil, app),
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{
						Target:  networkName,
						Aliases: []string{app.Key},
					},
				},
				// See DefaultLogDriver for why this is not `local`.
				LogDriver: appservice.DefaultLogDriver(),
			},
		},
	}
}

func defaultAppSettings(app *entity.App, timeNow time.Time) []*entity.Setting {
	// Init empty routing settings
	routingSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		Type:        base.SettingTypeAppRouting,
		Status:      base.SettingStatusActive,
		ObjectID:    app.ID,
		Inheritable: true,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	routingSetting.MustSetData(&entity.AppRoutingSettings{})

	// Init feature settings
	featureSettings := &entity.AppFeatureSettings{}
	entity.InitAppFeatureSettingsDefault(featureSettings)
	featureSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		Type:        base.SettingTypeAppFeatures,
		Status:      base.SettingStatusActive,
		ObjectID:    app.ID,
		Inheritable: true,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	featureSetting.MustSetData(featureSettings)

	return []*entity.Setting{routingSetting, featureSetting}
}

// replaceSettingsByType keeps each default no configured setting has the type
// of, then appends the configured settings.
func replaceSettingsByType(defaults, configured []*entity.Setting) []*entity.Setting {
	out := make([]*entity.Setting, 0, len(defaults)+len(configured))
	for _, setting := range defaults {
		if !slices.ContainsFunc(configured, func(c *entity.Setting) bool { return c.Type == setting.Type }) {
			out = append(out, setting)
		}
	}
	return append(out, configured...)
}

func appTags(app *entity.App, tags []string) []*entity.Tag {
	out := make([]*entity.Tag, 0, len(tags))
	for index, tag := range tags {
		out = append(out, &entity.Tag{ObjectID: app.ID, Tag: tag, Index: index})
	}
	return out
}
