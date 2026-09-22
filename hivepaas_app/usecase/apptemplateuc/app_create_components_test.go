package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// renderedWithComponents is what the service returns for a template that creates
// several apps: a gateway that owns the domain, an authentication service behind
// it, and one database shared by both.
func renderedWithComponents(t *testing.T, fakes *createFakes) *apptemplateservice.RenderResp {
	t.Helper()
	database := fakes.templates.resp
	stack := *database
	stack.Template = &templatemodel.Template{Metadata: templatemodel.Metadata{Name: "stack", Title: "Stack"}}
	stack.Dependencies = []*apptemplateservice.RenderedDependency{
		{Name: "db", AppName: "shop-db", Render: database},
	}
	stack.Components = []*apptemplateservice.RenderedComponent{
		{Name: "auth", Title: "Auth", AppName: "shop-auth", Result: database.Result},
		{Name: "gw", Title: "Gateway", AppName: "shop", Primary: true, Result: database.Result},
	}
	return &stack
}

// The primary component is the app the person named: same name, and the one the
// others are created for. The rest carry its id as their logical parent, which
// is what nests them under it everywhere apps are listed - and what makes
// appservice delete them with it, since it deletes an app's logical children.
func TestPlanAppsMakesThePrimaryComponentTheAppTheRequestNamed(t *testing.T) {
	_, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "shop"

	apps := planApps(req, renderedWithComponents(t, fakes))

	assert.Len(t, apps, 3, "one database, one secondary component, and the app itself")
	db, auth, gateway := apps[0], apps[1], apps[2]
	assert.Equal(t, "shop-db", db.name)
	assert.Equal(t, "shop-auth", auth.name)
	assert.Equal(t, "shop", gateway.name, "the primary component is the app that was named")
	assert.Equal(t, gateway.id, auth.links.createdForAppID)
	assert.Equal(t, gateway.id, auth.logicalParentID)
	assert.Empty(t, gateway.logicalParentID, "the primary is nobody's child")
}

// Every app of a stack records which part of it it is, and the primary records
// the others - the same two directions a dependency is recorded in.
func TestPlanAppsRecordsWhichComponentEachAppIs(t *testing.T) {
	_, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "shop"

	apps := planApps(req, renderedWithComponents(t, fakes))

	auth, gateway := apps[1], apps[2]
	assert.Equal(t, "auth", auth.links.component)
	assert.Equal(t, "gw", gateway.links.component)
	assert.Equal(t, []entity.AppTemplateComponent{{Name: "auth", AppID: auth.id}}, gateway.links.components)
	assert.Empty(t, auth.links.components, "only the primary lists the others")
}

// The secondary components are created before the primary, and the dependencies
// before all of them.
func TestProvisionAllCreatesComponentsBeforeTheAppTheyServe(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "shop"

	_, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithComponents(t, fakes)))

	assert.NoError(t, err)
	assert.Equal(t, []string{"shop-db", "shop-auth", "shop"}, fakes.provision.names)
}

// What is recorded on each app is what phase 2 will merge against, so every
// component carries its own binding rather than the primary carrying one for all
// of them.
func TestProvisionAllRecordsABindingOnEveryComponent(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "shop"

	provisioned, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithComponents(t, fakes)))

	assert.NoError(t, err)
	auth, gateway := provisioned.Apps[1].App, provisioned.Apps[2].App
	assert.Equal(t, "auth", binding(t, auth).Component)
	assert.Equal(t, "gw", binding(t, gateway).Component)
	assert.Equal(t, gateway.ID, binding(t, auth).CreatedForAppID)
	assert.Equal(t, auth.ID, binding(t, gateway).Components[0].AppID)
}
