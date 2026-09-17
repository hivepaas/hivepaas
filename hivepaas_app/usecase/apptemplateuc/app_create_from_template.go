package apptemplateuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

const (
	routingApplyRetryMax   = 3
	routingApplyRetryDelay = 500 * time.Millisecond
)

func (uc *UC) CreateAppFromTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
) (resp *apptemplatedto.CreateAppFromTemplateResp, err error) {
	// Before the transaction: rendering may reach GitHub, and nothing is locked
	// while it does. Every app of the request is rendered here, before any exists.
	rendered, err := uc.appTemplateService.Render(ctx, &apptemplateservice.RenderReq{
		Name:             req.Template,
		AppName:          req.Name,
		Version:          req.Version,
		Variant:          req.Variant,
		Params:           req.Params,
		DependencyParams: req.DependencyParams,
		ImageOverride:    req.ImageOverride,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	apps := planApps(req, rendered)
	for _, target := range apps {
		if doc := target.rendered.Result.Doc; doc.Deployment == nil || doc.Deployment.Source == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: the template deploys no image", target.rendered.Template.Metadata.Name)
		}
	}

	var created []*createdFromTemplate
	committed := false
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		// The swarm services are created inside the transaction but are not part
		// of it: a rolled-back request would leave them running with nothing
		// recorded.
		if err != nil && !committed {
			if removeErr := uc.removeServices(context.WithoutCancel(ctx), created); removeErr != nil {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		created, txErr = uc.provisionAll(ctx, db, auth, req, apps)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. The dependencies' deployments go first.
	for _, one := range created {
		if err = uc.taskQueue.ScheduleTask(ctx, one.deploymentTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return transformCreated(created), nil
}

// transformCreated describes what a request created: the app it asked for is the
// one without a role, and the others are its dependencies.
func transformCreated(created []*createdFromTemplate) *apptemplatedto.CreateAppFromTemplateResp {
	data := &apptemplatedto.CreateAppFromTemplateDataResp{
		Dependencies: make([]*apptemplatedto.CreatedDependencyResp, 0, len(created)),
	}
	for _, one := range created {
		app := &basedto.ObjectIDResp{ID: one.app.ID}
		deployment := &basedto.ObjectIDResp{ID: one.deployment.ID}
		if one.role == "" {
			data.App, data.Deployment = app, deployment
			continue
		}
		data.Dependencies = append(data.Dependencies,
			&apptemplatedto.CreatedDependencyResp{Name: one.role, App: app, Deployment: deployment})
	}
	return &apptemplatedto.CreateAppFromTemplateResp{Data: data}
}

type createdFromTemplate struct {
	// role is the dependency's name, empty for the app the request asked for.
	role           string
	app            *entity.App
	deployment     *entity.Deployment
	deploymentTask *entity.Task
}

// appToProvision is one app of a creation request: the one it asked for, or a
// dependency created for it.
type appToProvision struct {
	id       string
	name     string
	role     string
	rendered *apptemplateservice.RenderResp
	links    appTemplateLinks
}

// appTemplateLinks are what an app's binding records about the other apps of the
// same request.
type appTemplateLinks struct {
	dependencies    []entity.AppTemplateDependency
	createdForAppID string
}

// planApps lists the apps a request creates, dependencies first. Their ids are
// chosen here, so that each binding can name the others before any exists.
func planApps(
	req *apptemplatedto.CreateAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
) []*appToProvision {
	main := &appToProvision{id: gofn.Must(ulid.NewStringULID()), name: req.Name, rendered: rendered}
	apps := make([]*appToProvision, 0, len(rendered.Dependencies)+1)
	for _, dep := range rendered.Dependencies {
		depApp := &appToProvision{
			id:       gofn.Must(ulid.NewStringULID()),
			name:     dep.AppName,
			role:     dep.Name,
			rendered: dep.Render,
			links:    appTemplateLinks{createdForAppID: main.id},
		}
		main.links.dependencies = append(main.links.dependencies, entity.AppTemplateDependency{
			Name: dep.Name, AppID: depApp.id, Template: dep.Render.Template.Metadata.Name,
		})
		apps = append(apps, depApp)
	}
	return append(apps, main)
}

// provisionAll provisions the apps of a request in order. It returns what it
// created even when it fails part way, so the caller can remove their services.
func (uc *UC) provisionAll(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) ([]*createdFromTemplate, error) {
	created := make([]*createdFromTemplate, 0, len(apps))
	for _, target := range apps {
		one, err := uc.provisionFromTemplate(ctx, db, auth, req, target)
		if one != nil {
			created = append(created, one)
		}
		if err != nil {
			if target.role != "" {
				return created, hperrors.Wrap(err).WithExtraDetail(
					"while creating %s, which the template creates for %s", target.name, req.Name)
			}
			return created, hperrors.Wrap(err)
		}
	}
	return created, nil
}

// removeServices removes the services of apps whose transaction did not commit,
// newest first. Their records were rolled back, so a service that cannot be
// removed is running with nothing recorded, and the error names it.
func (uc *UC) removeServices(ctx context.Context, created []*createdFromTemplate) error {
	var errs []error
	for i := len(created) - 1; i >= 0; i-- {
		app := created[i].app
		if app == nil || app.ServiceID == "" {
			continue
		}
		err := uc.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		if err != nil {
			errs = append(errs, hperrors.Wrap(err).WithExtraDetail(
				"%s was created and could not be removed: remove service %s by hand", app.Name, app.ServiceID))
		}
	}
	return errors.Join(errs...)
}

// provisionFromTemplate is everything that happens inside the transaction. It
// returns what it created even when it fails part way, so the caller can remove
// the swarm service a rolled-back app would otherwise leave behind.
func (uc *UC) provisionFromTemplate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	target *appToProvision,
) (*createdFromTemplate, error) {
	timeNow := timeutil.NowUTC()
	rendered := target.rendered
	provisioned, err := uc.appProvisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
		ProjectID:    req.ProjectID,
		ProjectEnvID: req.ProjectEnvID,
		AppID:        target.id,
		Name:         target.name,
		Status:       base.AppStatusActive,
		Configure: func(configureCtx context.Context, configureDB database.IDB, app *entity.App,
			spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
			built, buildErr := uc.specService.BuildApp(configureCtx, configureDB, &specservice.BuildAppReq{
				App:     app,
				Doc:     rendered.Result.Doc,
				Spec:    spec,
				TimeNow: timeNow,
			})
			if buildErr != nil {
				return nil, hperrors.Wrap(buildErr)
			}
			binding, bindErr := newAppTemplateSetting(app, rendered, target.links, timeNow)
			if bindErr != nil {
				return nil, hperrors.Wrap(bindErr)
			}
			return append(built.Settings, binding), nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	created := &createdFromTemplate{role: target.role, app: provisioned.App}
	app := created.app

	if err = uc.applyEnvVars(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err).WithExtraDetail("while applying environment variables")
	}
	if err = uc.applyRouting(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err).WithExtraDetail("while applying routing settings")
	}
	if err = uc.createFirstDeployment(ctx, db, auth, created); err != nil {
		return created, hperrors.Wrap(err)
	}
	return created, uc.recordCreateFromTemplate(ctx, db, auth, app, rendered, target.links)
}

// applyEnvVars builds and applies the environment of every app in the new app's
// scope, as appcloneservice does after persisting a clone: the app kind makes
// shared variables such as HIVEPAAS_PASSWORD, which the app's own variables refer to.
func (uc *UC) applyEnvVars(ctx context.Context, db database.IDB, app *entity.App) error {
	// In a transaction: no nested transactions, and no concurrency.
	appEnvData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, app.GetObjectScope(),
		false, nil, false, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, db, appEnvData, false, false)
	for _, applyErr := range errMap {
		return hperrors.Wrap(applyErr)
	}
	return nil
}

// applyRouting writes the app's routing settings to traefik and to the service.
//
// It retries, which applying routing elsewhere does not have to: the service was
// created moments ago and swarm's own allocator is still writing to it, so an
// update carrying the version an inspect has just returned comes back as "update
// out of sequence" - about one create in three on a developer machine. Each
// attempt re-inspects the service and writes the same settings, so repeating it
// changes nothing beyond the version it carries.
func (uc *UC) applyRouting(ctx context.Context, db database.IDB, app *entity.App) error {
	routingSetting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppRouting)
	if routingSetting == nil {
		return nil
	}
	routingSettings, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	for attempt := range routingApplyRetryMax + 1 {
		if attempt > 0 {
			timer := time.NewTimer(routingApplyRetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return hperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}
		_, err = uc.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
			App:             app,
			RoutingSettings: routingSettings,
			RefObjects:      entity.NewRefObjects(),
		})
		if err == nil {
			return nil
		}
	}
	return hperrors.Wrap(err)
}

// createFirstDeployment queues the deployment that replaces the placeholder
// service image with the template's.
func (uc *UC) createFirstDeployment(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	created *createdFromTemplate,
) error {
	app := created.app
	deploymentSetting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("the template deploys no image")
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings)
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceAPI,
		SourceID: auth.User.ID,
	}
	err = uc.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingDeployments: []*entity.Deployment{deployment},
		UpsertingTasks:       []*entity.Task{task},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	created.deployment, created.deploymentTask = deployment, task
	return nil
}

// newAppTemplateSetting records the template an app was provisioned from. Secret
// parameters are stored encrypted, and the rendered base holds none of them.
func newAppTemplateSetting(
	app *entity.App,
	rendered *apptemplateservice.RenderResp,
	links appTemplateLinks,
	timeNow time.Time,
) (*entity.Setting, error) {
	result := rendered.Result
	data := &entity.AppTemplateSettings{
		Source:   rendered.Source,
		Template: rendered.Template.Metadata.Name,
		Title:    rendered.Template.Metadata.Title,
		Version:  result.Version.Name,
		Params:   make(map[string]*entity.AppTemplateParam, len(result.Params)),
		Base: entity.AppTemplateBase{
			Revision:       rendered.Revision,
			Release:        result.Version.Release,
			TemplateSHA256: rendered.Entry.File.SHA256,
			Rendered:       string(result.Base),
			RenderedSHA256: result.BaseSHA256,
			AppliedAt:      timeNow,
		},
	}
	if result.Variant != nil {
		data.Variant = result.Variant.Name
	}
	data.ImageOverride = result.ImageOverride
	data.Dependencies = links.dependencies
	data.CreatedForAppID = links.createdForAppID
	for name, value := range result.Params {
		if value.Value == nil {
			continue
		}
		if value.Param.Type == templatemodel.ParamTypeSecret {
			data.Params[name] = &entity.AppTemplateParam{Secret: entity.NewEncryptedField(value.Text())}
			continue
		}
		data.Params[name] = &entity.AppTemplateParam{Value: value.Text()}
	}

	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeAppTemplate,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentAppTemplateSettingsVersion,
		UpdateVer: 1,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := setting.SetData(data); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}
