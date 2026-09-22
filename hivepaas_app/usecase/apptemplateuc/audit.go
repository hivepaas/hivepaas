package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordCreateFromTemplate records the creation as app-create, like any other,
// with where the app came from. Parameter values are never recorded: some are
// secrets, and the rest are on the app itself.
func (uc *UC) recordCreateFromTemplate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	target *appToProvision,
	links appTemplateLinks,
) error {
	rendered := target.rendered
	result := target.result
	detail := auditdetail.New().
		Set("projectId", app.ProjectID).
		Set("envId", app.ProjectEnvID).
		Set("serviceId", app.ServiceID).
		Set("source", rendered.Source).
		Set("template", rendered.Template.Metadata.Name).
		Set("version", result.Version.Name).
		Set("release", result.Version.Release).
		Set("revision", rendered.Revision)
	if result.Variant != nil {
		detail.Set("variant", result.Variant.Name)
	}
	if result.ImageOverride != "" {
		detail.Set("imageOverride", result.ImageOverride)
	}
	if links.component != "" {
		detail.Set("component", links.component)
	}
	if len(links.components) > 0 {
		components := make([]map[string]string, 0, len(links.components))
		for _, component := range links.components {
			components = append(components, map[string]string{"name": component.Name, "appId": component.AppID})
		}
		detail.Set("components", components)
	}
	if len(links.dependencies) > 0 {
		deps := make([]map[string]string, 0, len(links.dependencies))
		for _, dep := range links.dependencies {
			deps = append(deps, map[string]string{"name": dep.Name, "appId": dep.AppID, "template": dep.Template})
		}
		detail.Set("dependencies", deps)
	}
	detail.Set("createdFor", links.createdForAppID)

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppCreate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPICreate,
		Section:  "create",
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.String(),
	})
	return hperrors.Wrap(err)
}
