package specserviceimpl

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const reportPath = "projects/project_a/envs/dev/apps/report"

// functionSourceBody is a function's deployment source as a bundle holds it,
// its code inline; what it leaves out is the runtime's default.
func functionSourceBody(runtime string) map[string]any {
	return map[string]any{"activeMethod": "function", "functionSource": map[string]any{
		"runtime":    runtime,
		"entrypoint": map[string]any{"file": "", "handler": ""},
		"code": map[string]any{"inline": map[string]any{"files": []any{
			map[string]any{"path": "index.js", "content": "export default () => ({})"},
		}}},
	}}
}

func kindBody(category string) map[string]any {
	return map[string]any{"category": category}
}

// addReport gives the bundle an app the target does not have, report, in the
// fixture's env.
func addReport(bundle *specmodel.ImportBundle, kind, source map[string]any, settings map[string]any) {
	const key = "report"
	if settings == nil {
		settings = map[string]any{}
	}
	if kind != nil {
		settings["kind"] = kind
	}
	doc := &specmodel.AppDoc{App: key, Name: key, Settings: settings}
	if source != nil {
		doc.Deployment = &specmodel.Deployment{Source: source}
	}
	bundle.Envs["project_a"]["dev"].Apps[key] = doc
}

func blockedCodes(node *specmodel.PlanNode) []string {
	var codes []string
	for _, issue := range node.Issues {
		if issue.Severity == specmodel.SeverityBlocked {
			codes = append(codes, issue.Code)
		}
	}
	return codes
}

// A function in a bundle is created as a function: its source normalized as
// creating one does it, its routing at its runtime's port, its first
// deployment when the options ask.
func TestImportCreatesAFunctionAsCreatingOneDoes(t *testing.T) {
	svc, bundle := planFixture(t)
	addReport(bundle, kindBody("function"), functionSourceBody("node24"), routingAt("report.example.com"))

	resp := apply(t, svc, bundle, applyReqWith(t, svc, bundle, specmodel.ImportOptions{DeployCreated: true}))

	provision := svc.appProvisionService.(*fakeProvisionService)
	var req = provision.reqs
	if !assert.Len(t, req, 1) {
		return
	}
	byType := map[base.SettingType]*entity.Setting{}
	for _, setting := range provision.settings[req[0].AppID] {
		byType[setting.Type] = setting
	}
	assert.True(t, entity.IsFunctionKind(byType[base.SettingTypeAppKind]))
	source := byType[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings().FunctionSource
	if assert.NotNil(t, source) {
		assert.Equal(t, base.FunctionContractV1, source.Contract)
		assert.Equal(t, entity.FunctionEntrypoint{File: "index.js", Handler: "default"}, source.Entrypoint)
		assert.Equal(t, base.FunctionMaxConcurrencyDefault, source.MaxConcurrency)
		assert.Equal(t, base.FunctionMaxBodySizeDefault, source.MaxBodySize)
		assert.EqualValues(t, base.FunctionTimeoutDefault, source.Timeout)
	}
	routing := byType[base.SettingTypeAppRouting].MustAsAppRoutingSettings()
	assert.Equal(t, base.FunctionPort, routing.Port, "a function answers at its runtime's port")
	if assert.Len(t, routing.Domains, 1) {
		assert.Equal(t, base.FunctionPort, routing.Domains[0].ContainerPort)
	}
	assert.Len(t, resp.Deployments, 1, "deployed as the options ask")
}

func TestImportBlocksAFunctionWhoseSourceIsNotValid(t *testing.T) {
	svc, bundle := planFixture(t)
	addReport(bundle, kindBody("function"), functionSourceBody("cobol"), nil)

	out := plan(t, svc, bundle, specmodel.ImportOptions{})

	report := node(t, out, reportPath)
	assert.Equal(t, []string{specmodel.CodeFunctionSourceInvalid}, blockedCodes(report))
	problems, _ := report.Issues[len(report.Issues)-1].Detail["problems"].([]string)
	assert.True(t, len(problems) > 0 && strings.Contains(strings.Join(problems, " "), "runtime"),
		"the problems name what is wrong: %v", problems)
}

// A kind and a source that disagree make neither an app nor a function.
func TestImportBlocksAKindThatDisagreesWithItsSource(t *testing.T) {
	cases := map[string][2]map[string]any{
		"a function kind with an image": {kindBody("function"), imageSource("nginx:1.27")},
		"a function source on a webapp": {kindBody("webapp"), functionSourceBody("node24")},
		"a function source and no kind": {nil, functionSourceBody("node24")},
	}
	for name, kindAndSource := range cases {
		svc, bundle := planFixture(t)
		addReport(bundle, kindAndSource[0], kindAndSource[1], nil)

		out := plan(t, svc, bundle, specmodel.ImportOptions{})

		assert.Equal(t, []string{specmodel.CodeFunctionKindMismatch}, blockedCodes(node(t, out, reportPath)), name)
	}
}

// An app does not change its kind: a function stays one, and an app does not
// become one.
func TestImportBlocksAnAppChangingKind(t *testing.T) {
	svc, bundle := planFixture(t)
	hello := bundle.Envs["project_a"]["dev"].Apps["hello"]
	hello.Settings["kind"] = kindBody("webapp")
	hello.Deployment.Source = imageSource("nginx:1.27")

	out := plan(t, svc, bundle, specmodel.ImportOptions{})

	assert.Equal(t, []string{specmodel.CodeAppKindChanged},
		blockedCodes(node(t, out, "projects/project_a/envs/dev/apps/hello")))

	svc, bundle = planFixture(t)
	frontend := bundle.Envs["project_a"]["dev"].Apps["frontend"]
	if frontend.Settings == nil {
		frontend.Settings = map[string]any{}
	}
	frontend.Settings["kind"] = kindBody("function")
	frontend.Deployment = &specmodel.Deployment{Source: functionSourceBody("node24")}

	out = plan(t, svc, bundle, specmodel.ImportOptions{})

	assert.Equal(t, []string{specmodel.CodeAppKindChanged},
		blockedCodes(node(t, out, "projects/project_a/envs/dev/apps/frontend")))
}

// A cluster of several nodes pulls a function's image from a registry, as it
// does a repository app's: a function created without one is blocked there.
func TestImportBlocksAFunctionWithoutARegistryOnSeveralNodes(t *testing.T) {
	svc, bundle := planFixture(t)
	svc.clusterService.(*fakeClusterService).multiNode = true
	addReport(bundle, kindBody("function"), functionSourceBody("node24"), nil)

	out := plan(t, svc, bundle, specmodel.ImportOptions{})

	assert.Equal(t, []string{specmodel.CodeFunctionBuildSource}, blockedCodes(node(t, out, reportPath)))
	assert.Empty(t, blockedCodes(node(t, out, "projects/project_a/envs/dev/apps/hello")),
		"a function whose source does not change is not asked again")
}

// What a plan blocks is not written, whichever path writes settings: a bundle
// written by hand cannot make a function past the checks.
func TestApplyRefusesAFunctionTheChecksBlock(t *testing.T) {
	svc, bundle := planFixture(t)
	addReport(bundle, nil, functionSourceBody("node24"), nil)
	req := applyReq(t, svc, bundle)

	_, err := svc.applyBundle(context.Background(), nil, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrSpecImportBlocked)
	assert.Empty(t, svc.appProvisionService.(*fakeProvisionService).reqs)
}

// functionCallJob is a function-invoke job as a bundle holds it, at an app.
func functionCallJob() map[string]any {
	return map[string]any{"nightly": map[string]any{
		"jobType":        "function-invoke",
		"functionInvoke": map[string]any{"method": "POST", "path": "/report"},
	}}
}

// A function's call lives in a function, as saving one requires: an app that
// is no function cannot be given one by an import.
func TestImportBlocksAFunctionCallOnAnAppThatIsNoFunction(t *testing.T) {
	svc, bundle := planFixture(t)
	addReport(bundle, kindBody("webapp"), imageSource("nginx:1.27"),
		map[string]any{"schedJobs": functionCallJob()})

	out := plan(t, svc, bundle, specmodel.ImportOptions{})

	assert.Equal(t, []string{specmodel.CodeFunctionCallOnApp}, blockedCodes(node(t, out, reportPath)))

	svc, bundle = planFixture(t)
	addReport(bundle, kindBody("function"), functionSourceBody("node24"),
		map[string]any{"schedJobs": functionCallJob()})

	out = plan(t, svc, bundle, specmodel.ImportOptions{})

	assert.Empty(t, blockedCodes(node(t, out, reportPath)), "a function's own call is imported with it")
}
