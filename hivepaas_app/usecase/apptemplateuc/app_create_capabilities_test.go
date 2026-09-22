package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// capabilityTemplateYAML asks for what opensearch actually needs: the memory it
// may lock, and a map count the kernel would otherwise refuse it.
const capabilityTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: search
  title: Search
  tagline: Test search engine
  description: Test.
  categories: [databases/search]
  icon: icons/search.svg
  requires: {versionCode: v000001}
versions:
  - {name: "3", release: "3.0", default: true, image: "opensearchproject/opensearch:3.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    resources:
      capabilities:
        capabilityAdd: [IPC_LOCK]
        sysctls: {vm.max_map_count: "262144"}
        ulimits: [{name: memlock, soft: -1, hard: -1}]
  settings:
    routing: {port: 9200}
`

type fakePermissionManager struct {
	permission.Manager
	granted bool
	checked []permission.AccessCheck
}

func (f *fakePermissionManager) CheckAccess(
	_ context.Context, _ database.IDB, _ *basedto.Auth, check permission.AccessCheck,
) (bool, error) {
	f.checked = append(f.checked, check)
	return f.granted, nil
}

func capabilityApps(t *testing.T) []*appToProvision {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(capabilityTemplateYAML))
	assert.NoError(t, err)
	result, err := templaterender.Render(&templaterender.Request{Template: tmpl})
	assert.NoError(t, err)
	return []*appToProvision{{
		id:     "app-1",
		name:   "search",
		result: result,
		rendered: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{Template: tmpl},
			Result:       result,
		},
	}}
}

func TestCheckCapabilitiesNeedsWriteOnTheClusterModule(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}

	err := uc.checkCapabilities(context.Background(), testAuth(), capabilityApps(t))

	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	var hpErr hperrors.HPError
	assert.ErrorAs(t, err, &hpErr)
	detail := hpErr.Build("en").Detail
	assert.Contains(t, detail, "search asks for IPC_LOCK")
	assert.Contains(t, detail, "sysctls")
	assert.Contains(t, detail, "Cluster")

	assert.Len(t, permissions.checked, 1)
	check, ok := permissions.checked[0].(*permission.ModuleAccessCheck)
	assert.True(t, ok)
	assert.Equal(t, base.ResourceModuleCluster, check.Module)
	assert.Equal(t, base.ActionTypeWrite, check.Action)
}

func TestCheckCapabilitiesPassesWhenTheCallerMayGrantThem(t *testing.T) {
	permissions := &fakePermissionManager{granted: true}
	uc := &UC{permissionManager: permissions}

	assert.NoError(t, uc.checkCapabilities(context.Background(), testAuth(), capabilityApps(t)))
}

// A template that asks for nothing is not a permission question, so nothing is
// asked - which is what keeps every ordinary template creatable by whoever may
// create apps.
func TestCheckCapabilitiesAsksNothingOfAnOrdinaryTemplate(t *testing.T) {
	uc, fakes := newCreateTest(t)
	permissions := &fakePermissionManager{granted: false}
	uc.permissionManager = permissions

	apps := planApps(testCreateReq(), fakes.templates.resp)

	assert.NoError(t, uc.checkCapabilities(context.Background(), testAuth(), apps))
	assert.Empty(t, permissions.checked)
}

// A dependency is created by the same request and runs on the same nodes, so it
// is gated the same way - and the message says which app asked.
func TestCheckCapabilitiesCoversTheDependenciesOfARequest(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}
	apps := capabilityApps(t)
	main := &appToProvision{id: "app-2", name: "main-app"}
	main.rendered = plainRender(t)
	main.result = main.rendered.Result
	apps[0].role, apps[0].logicalParentID = "search", main.id

	err := uc.checkCapabilities(context.Background(), testAuth(), append(apps, main))

	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	var hpErr hperrors.HPError
	assert.ErrorAs(t, err, &hpErr)
	assert.Contains(t, hpErr.Build("en").Detail, "search asks for IPC_LOCK")
	assert.NotContains(t, hpErr.Build("en").Detail, "main-app", "the app that asks for nothing is not named")
}

// plainRender is a rendered document that asks for no capabilities.
func plainRender(t *testing.T) *apptemplateservice.RenderResp {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(testTemplateYAML))
	assert.NoError(t, err)
	result, err := templaterender.Render(&templaterender.Request{
		Template: tmpl,
		Params:   map[string]any{"dataVolume": "vol-1"},
	})
	assert.NoError(t, err)
	assert.Nil(t, capabilitiesOf(result.Doc))
	return &apptemplateservice.RenderResp{
		TemplateResp: apptemplateservice.TemplateResp{Template: tmpl},
		Result:       result,
	}
}

func capabilitiesOf(doc *specmodel.AppDoc) *specmodel.Capabilities {
	if doc == nil || doc.Deployment == nil || doc.Deployment.Resources == nil {
		return nil
	}
	return doc.Deployment.Resources.Capabilities
}
