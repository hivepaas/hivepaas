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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
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
		ImageTag:         req.ImageTag,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	apps := planApps(req, rendered)
	for _, target := range apps {
		if doc := target.result.Doc; doc.Deployment == nil || doc.Deployment.Source == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: the template deploys no image", target.rendered.Template.Metadata.Name)
		}
	}
	if err = uc.checkAppRefs(ctx, req, rendered); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.checkCapabilities(ctx, auth, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.checkDockerAPI(ctx, auth, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.checkSharedMounts(ctx, auth, req, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.checkPublishedPorts(ctx, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.checkDomains(ctx, req, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Whatever a previous install of these apps left on the volumes goes first,
	// and only when the request says so. The preflight endpoint is where anybody
	// is told there is something to clear; this is where it is cleared.
	if req.ResetStorage {
		plan, planErr := uc.planStorage(ctx, uc.db, req, apps)
		if planErr != nil {
			return nil, hperrors.Wrap(planErr)
		}
		if err = uc.volumeService.RemoveAppStoragePaths(ctx, uc.db, plan.Request); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	var provisioned *appprovisionservice.ProvisionAppsResp
	committed := false
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		// The swarm services, secrets and configs are created inside the
		// transaction but are not part of it: a rolled-back request would leave
		// them behind with nothing recorded.
		if err != nil && !committed && provisioned != nil {
			if cleanupErr := provisioned.Cleanup(context.WithoutCancel(ctx)); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		provisioned, txErr = uc.provisionAll(ctx, db, auth, req, apps)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. The dependencies' deployments go first.
	for _, one := range provisioned.Apps {
		if err = uc.taskQueue.ScheduleTask(ctx, one.DeploymentTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
		if err = uc.taskQueue.ScheduleTask(ctx, one.CertTasks...); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return transformCreated(apps, provisioned.Apps), nil
}

// checkDomains refuses a request whose domains cannot be served before anything
// is created for it.
//
// These are the two checks the app's routing settings run - the project allows
// the domain, and no other app holds it - plus one this request needs and they
// do not: two apps created together must not ask for the same address. Running
// them here rather than while provisioning is what keeps a request that names a
// taken domain from leaving a database app behind.
func (uc *UC) checkDomains(
	ctx context.Context,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) error {
	claimed := map[string]string{}
	var domains []string
	for _, target := range apps {
		// A routing block that does not read has no domains to check here:
		// building it refuses it, before anything is created.
		rendered, _ := specmodel.ActiveDomains(target.result.Doc)
		for _, domain := range rendered {
			if by, taken := claimed[domain]; taken {
				return hperrors.Wrap(hperrors.ErrDomainInUse).WithParam("Domain", domain).
					WithExtraDetail("%s and %s are created together and ask for the same address", by, target.name)
			}
			claimed[domain] = target.name
			domains = append(domains, domain)
		}
	}
	if len(domains) == 0 {
		return nil
	}
	if err := uc.domainService.VerifyProjectDomains(ctx, uc.db, req.ProjectID, domains); err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.domainService.VerifyDomainsAvailable(ctx, uc.db, domains, nil))
}

// provisionAll is everything that happens inside the transaction: the apps of
// the request are created, and each creation is recorded.
//
// It returns what provisioning made even when it fails part way, so that the
// caller can undo in docker what the transaction cannot take back.
func (uc *UC) provisionAll(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) (*appprovisionservice.ProvisionAppsResp, error) {
	provisioned, err := uc.appProvisionService.ProvisionApps(ctx, db,
		&appprovisionservice.ProvisionAppsReq{Apps: uc.provisionRequests(auth, req, apps)})
	if err != nil {
		return provisioned, hperrors.Wrap(err)
	}
	return provisioned, uc.recordCreated(ctx, db, auth, apps, provisioned.Apps)
}

// provisionRequests turns the plan into what provisioning is asked for. The apps
// are provisioned in this order, so a dependency exists before the app that was
// created with it.
func (uc *UC) provisionRequests(
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) []*appprovisionservice.ProvisionAppReq {
	timeNow := timeutil.NowUTC()
	out := make([]*appprovisionservice.ProvisionAppReq, 0, len(apps))
	for _, target := range apps {
		out = append(out, &appprovisionservice.ProvisionAppReq{
			ProjectID:    req.ProjectID,
			ProjectEnvID: req.ProjectEnvID,
			AppID:        target.id,
			Name:         target.name,
			Status:       base.AppStatusActive,
			// A dependency belongs to the app it was created with, which is what
			// keeps it out of a listing of apps and takes it along when that app
			// is deleted.
			LogicalParentID: target.logicalParentID,
			Configure:       uc.configureFromTemplate(target, timeNow),
			Deployment: &appprovisionservice.FirstDeployment{
				Source:   base.DeploymentTriggerSourceAPI,
				SourceID: auth.User.ID,
			},
		})
	}
	return out
}

// configureFromTemplate builds an app's settings from the document rendered for
// it, and adds the binding that records where the app came from.
func (uc *UC) configureFromTemplate(
	target *appToProvision,
	timeNow time.Time,
) appprovisionservice.ConfigureFunc {
	return func(ctx context.Context, db database.IDB, app *entity.App,
		spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
		built, err := uc.specService.BuildApp(ctx, db, &specservice.BuildAppReq{
			App:     app,
			Doc:     target.result.Doc,
			Spec:    spec,
			TimeNow: timeNow,
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		binding, err := newAppTemplateSetting(app, target.rendered, target.result, target.links, timeNow)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return append(built.Settings, binding), nil
	}
}

// recordCreated audits each app of the request, in the order they were created.
func (uc *UC) recordCreated(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	apps []*appToProvision,
	created []*appprovisionservice.ProvisionAppResp,
) error {
	for i, one := range created {
		if err := uc.recordCreateFromTemplate(ctx, db, auth, one.App, apps[i], apps[i].links); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// transformCreated describes what a request created: the app it asked for is the
// one without a role - the primary component, for a template that has them - and
// the others are its components and its dependencies, told apart by whether the
// template declared them as one.
func transformCreated(
	apps []*appToProvision,
	created []*appprovisionservice.ProvisionAppResp,
) *apptemplatedto.CreateAppFromTemplateResp {
	data := &apptemplatedto.CreateAppFromTemplateDataResp{
		Dependencies: make([]*apptemplatedto.CreatedDependencyResp, 0, len(created)),
		Components:   make([]*apptemplatedto.CreatedComponentResp, 0, len(created)),
	}
	for i, one := range created {
		app := &basedto.ObjectIDResp{ID: one.App.ID}
		deployment := &basedto.ObjectIDResp{ID: one.Deployment.ID}
		switch {
		case apps[i].role == "":
			data.App, data.Deployment = app, deployment
		case apps[i].links.component != "":
			data.Components = append(data.Components,
				&apptemplatedto.CreatedComponentResp{Name: apps[i].role, App: app, Deployment: deployment})
		default:
			data.Dependencies = append(data.Dependencies,
				&apptemplatedto.CreatedDependencyResp{Name: apps[i].role, App: app, Deployment: deployment})
		}
	}
	return &apptemplatedto.CreateAppFromTemplateResp{Data: data}
}

// appToProvision is one app of a creation request: the one it asked for, or a
// dependency created for it.
type appToProvision struct {
	id   string
	name string
	role string
	// logicalParentID is the app this one is created to serve, empty for the app
	// the request is about.
	logicalParentID string
	rendered        *apptemplateservice.RenderResp
	// result is this app's own render. It is rendered.Result for a dependency and
	// for a template with one app; a template with components renders one of
	// these per component, and they differ in every way that matters.
	result *templaterender.Result
	links  appTemplateLinks
}

// key is the name this app will answer to on the project's network, and the
// directory its storage gets inside a volume. It is decided by the name, so it
// is known before the app exists.
func (a *appToProvision) key() string {
	return projecthelper.CalcAppKey(a.name)
}

// appTemplateLinks are what an app's binding records about the other apps of the
// same request.
type appTemplateLinks struct {
	dependencies    []entity.AppTemplateDependency
	components      []entity.AppTemplateComponent
	createdForAppID string
	// component is the role this app plays in a template that creates several,
	// empty for an app that is its template's only one.
	component string
}

// planApps lists the apps a request creates, dependencies first. Their ids are
// chosen here, so that each binding can name the others before any exists.
func planApps(
	req *apptemplatedto.CreateAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
) []*appToProvision {
	main := &appToProvision{
		id: gofn.Must(ulid.NewStringULID()), name: req.Name, rendered: rendered, result: rendered.Result,
	}
	apps := make([]*appToProvision, 0, len(rendered.Dependencies)+len(rendered.Components)+1)
	for _, dep := range rendered.Dependencies {
		depApp := &appToProvision{
			id:              gofn.Must(ulid.NewStringULID()),
			name:            dep.AppName,
			role:            dep.Name,
			rendered:        dep.Render,
			result:          dep.Render.Result,
			logicalParentID: main.id,
			links:           appTemplateLinks{createdForAppID: main.id},
		}
		main.links.dependencies = append(main.links.dependencies, entity.AppTemplateDependency{
			Name: dep.Name, AppID: depApp.id, Template: dep.Render.Template.Metadata.Name,
		})
		apps = append(apps, depApp)
	}
	return append(apps, planComponents(rendered, main)...)
}

// planComponents lists the component apps of a template that creates several,
// in the order its needs put them, with the primary one last.
//
// The primary app is the one the person named, so it is main itself: the same
// id, the same name, and the render of the primary component. The rest are
// created before it and carry its id as their logical parent, which is what
// makes them nest under it everywhere apps are listed.
func planComponents(rendered *apptemplateservice.RenderResp, main *appToProvision) []*appToProvision {
	if len(rendered.Components) == 0 {
		return []*appToProvision{main}
	}
	apps := make([]*appToProvision, 0, len(rendered.Components))
	for _, component := range rendered.Components {
		if component.Primary {
			main.links.component = component.Name
			main.result = component.Result
			continue
		}
		componentApp := &appToProvision{
			id:              gofn.Must(ulid.NewStringULID()),
			name:            component.AppName,
			role:            component.Name,
			rendered:        rendered,
			result:          component.Result,
			logicalParentID: main.id,
			links: appTemplateLinks{
				createdForAppID: main.id,
				component:       component.Name,
			},
		}
		main.links.components = append(main.links.components, entity.AppTemplateComponent{
			Name: component.Name, AppID: componentApp.id,
		})
		apps = append(apps, componentApp)
	}
	return append(apps, main)
}

// newAppTemplateSetting records the template an app was provisioned from. Secret
// parameters are stored encrypted, and the rendered base holds none of them.
func newAppTemplateSetting(
	app *entity.App,
	rendered *apptemplateservice.RenderResp,
	result *templaterender.Result,
	links appTemplateLinks,
	timeNow time.Time,
) (*entity.Setting, error) {
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
	data.Components = links.components
	data.Component = links.component
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
