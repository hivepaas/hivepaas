package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

const workerPath = "projects/project_a/envs/dev/apps/worker"

// addWorker gives the bundle an app the target does not have, in an env it
// does, mounting its own directory of the project's volume.
func addWorker(bundle *specmodel.ImportBundle, settings map[string]any) {
	bundle.Envs["project_a"]["dev"].Apps["worker"] = &specmodel.AppDoc{
		App: "worker", Name: "Worker", Settings: settings,
		Deployment: &specmodel.Deployment{Storage: &specmodel.Storage{Mounts: map[string]specmodel.Mount{
			"/data": {Type: mount.TypeVolume, Source: "projects/project_a/volumes/default",
				VolumeOptions: &specmodel.VolumeOptions{Subpath: "data"}},
		}}},
	}
}

func routingAt(domain string) map[string]any {
	return map[string]any{"routing": map[string]any{
		"port": 80, "exposePublicly": true,
		"domains": []any{map[string]any{"enabled": true, "domain": domain, "protocol": "http"}},
	}}
}

func issuesOf(node *specmodel.PlanNode, code string) []specmodel.Issue {
	var out []specmodel.Issue
	for _, issue := range node.Issues {
		if issue.Code == code {
			out = append(out, issue)
		}
	}
	return out
}

func TestPlanReportsADomainAnotherAppHolds(t *testing.T) {
	svc, bundle := planFixture(t)
	svc.domainService.(*fakeDomainService).held = map[string]string{"api.example.com": "app_other"}
	backendRouting(bundle)["exposePublicly"] = true

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	if issues := issuesOf(backend, specmodel.CodeDomainInUse); assert.Len(t, issues, 1) {
		assert.Equal(t, specmodel.SeverityFixable, issues[0].Severity)
		assert.Equal(t, map[string]any{"domain": "api.example.com"}, issues[0].Detail)
	}
}

// The app the import rewrites holds its own domain: that is not a conflict.
func TestPlanDoesNotCountADomainTheRewrittenAppHolds(t *testing.T) {
	svc, bundle := planFixture(t)
	svc.domainService.(*fakeDomainService).held = map[string]string{"api.example.com": "app_1"}
	backendRouting(bundle)["exposePublicly"] = true

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	assert.Empty(t, backend.Issues)
}

func TestPlanReportsTwoImportedAppsAskingForOneDomain(t *testing.T) {
	svc, bundle := planFixture(t)
	backendRouting(bundle)["exposePublicly"] = true
	addWorker(bundle, routingAt("api.example.com"))

	worker := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), workerPath)

	if issues := issuesOf(worker, specmodel.CodeDomainInUse); assert.Len(t, issues, 1) {
		assert.Equal(t, map[string]any{"domain": "api.example.com", "with": backendPath}, issues[0].Detail)
	}
}

func TestPlanReportsAPortAnotherServiceHolds(t *testing.T) {
	for name, holder := range map[string]string{"another service": "svc_other", "the app's own": "svc_1"} {
		t.Run(name, func(t *testing.T) {
			svc, bundle := planFixture(t)
			svc.clusterService.(*fakeClusterService).ports = map[clusterservice.PortRef]string{
				{Published: 8080, Protocol: network.TCP}: holder,
			}
			bundle.Envs["project_a"]["dev"].Apps["backend"].Deployment.Networks.EndpointSpec = &specmodel.EndpointSpec{
				Ports: []*specmodel.PortConfig{{Target: 80, Published: 8080}},
			}

			backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

			issues := issuesOf(backend, specmodel.CodePortInUse)
			if holder == "svc_1" {
				assert.Empty(t, issues)
				return
			}
			if assert.Len(t, issues, 1) {
				assert.Equal(t, map[string]any{"port": "8080/tcp"}, issues[0].Detail)
			}
		})
	}
}

func TestPlanReportsAVolumePinnedToANodeTheTargetLacks(t *testing.T) {
	for nodeID, wantIssue := range map[string]bool{"node_gone": true, "node_1": false} {
		t.Run(nodeID, func(t *testing.T) {
			svc, bundle := planFixture(t)
			volumes, _ := bundle.Projects["project_a"].Settings["volumes"].(map[string]any)
			volumes["default"].(map[string]any)["nodeId"] = nodeID

			settings := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), "projects/project_a/settings")

			issues := issuesOf(settings, specmodel.CodeNodeNotFound)
			if !wantIssue {
				assert.Empty(t, issues)
				return
			}
			if assert.Len(t, issues, 1) {
				assert.Equal(t, map[string]any{"setting": "volumes/default", "node": nodeID}, issues[0].Detail)
			}
		})
	}
}

// An app created on storage that already holds something is warned about, and
// so is one whose storage could not be looked at: unseen is not empty.
func TestPlanWarnsAboutStorageACreatedAppWouldStartOn(t *testing.T) {
	for code, state := range map[string]*volumeservice.AppStorageState{
		specmodel.CodeStorageNotEmpty:  {Checked: true, Exists: true, VolumeName: "default", Path: "p/dev/worker/data"},
		specmodel.CodeStorageUnchecked: {VolumeName: "default", Path: "p/dev/worker/data"},
	} {
		t.Run(code, func(t *testing.T) {
			svc, bundle := planFixture(t)
			svc.volumeService.(*fakeExportVolumeService).storage = map[string]*volumeservice.AppStorageState{
				"vol_setting_1": state,
			}
			addWorker(bundle, nil)

			worker := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), workerPath)

			if issues := issuesOf(worker, code); assert.Len(t, issues, 1) {
				assert.Equal(t, specmodel.SeverityWarning, issues[0].Severity)
				assert.Equal(t, map[string]any{"volume": "default", "path": "p/dev/worker/data"}, issues[0].Detail)
			}
		})
	}
}

func TestPlanOfACreatedAppOnEmptyStorageWarnsOfNothing(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, nil)

	assert.Empty(t, node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), workerPath).Issues)
}

func planWith(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, req *specservice.ValidateImportReq,
) *specmodel.ImportPlan {
	t.Helper()
	req.Scope = entity.NewObjectScopeGlobal()
	req.Options.Existing = specmodel.ExistingUpdate
	out, err := svc.planImport(context.Background(), nil, req, bundle)
	assert.NoError(t, err)
	return out
}

func TestPlanSkipsAnAppGrantingCapabilitiesTheOperatorMayNot(t *testing.T) {
	for allowed, wantSkip := range map[bool]bool{false: true, true: false} {
		svc, bundle := planFixture(t)
		bundle.Envs["project_a"]["dev"].Apps["backend"].Deployment.Resources.Capabilities =
			&specmodel.Capabilities{CapabilityAdd: []string{"NET_ADMIN"}}
		asked := 0

		backend := node(t, planWith(t, svc, bundle, &specservice.ValidateImportReq{
			MayWriteCluster: func(context.Context) (bool, error) { asked++; return allowed, nil },
		}), backendPath)

		assert.Equal(t, 1, asked)
		if !wantSkip {
			assert.Equal(t, specmodel.ActionUpdate, backend.Action)
			continue
		}
		assert.Equal(t, specmodel.ActionSkip, backend.Action)
		if issues := issuesOf(backend, specmodel.CodeCapabilityNotPermitted); assert.Len(t, issues, 1) {
			assert.Equal(t, map[string]any{"capabilities": []string{"NET_ADMIN"}}, issues[0].Detail)
		}
	}
}

// A mount into the directory of an app the import leaves alone needs Write on
// that app; one into an app the import writes does not.
func TestPlanSkipsAnAppReachingStorageTheOperatorMayNotWrite(t *testing.T) {
	for name, selectors := range map[string][]string{
		"the other app is left alone": {backendPath},
		"the other app is imported":   {backendPath, "projects/project_a/envs/dev/apps/frontend"},
	} {
		t.Run(name, func(t *testing.T) {
			svc, bundle := planFixture(t)
			storage := backendStorage(bundle)
			data := storage.Mounts["/var/lib/postgresql/data"]
			data.SourceApp = &specmodel.MountSourceApp{App: "frontend", Write: true}
			storage.Mounts["/var/lib/postgresql/data"] = data
			bundle.Envs["project_a"]["dev"].Apps["frontend"].Settings = map[string]any{
				"envVars": map[string]any{"data": []any{map[string]any{"k": "MODE", "v": "import"}}},
			}
			selection := specmodel.Selection{Include: selectors}
			var askedFor []string

			backend := node(t, planWith(t, svc, bundle, &specservice.ValidateImportReq{
				Selection: selection,
				MayWriteApp: func(_ context.Context, app *entity.App) (bool, error) {
					askedFor = append(askedFor, app.Key)
					return false, nil
				},
			}), backendPath)

			if len(selectors) == 2 {
				assert.Empty(t, askedFor)
				assert.Equal(t, specmodel.ActionUpdate, backend.Action)
				return
			}
			assert.Equal(t, []string{"frontend"}, askedFor)
			assert.Equal(t, specmodel.ActionSkip, backend.Action)
			if issues := issuesOf(backend, specmodel.CodeSharedMountNotPermitted); assert.Len(t, issues, 1) {
				assert.Equal(t, map[string]any{"mount": "/var/lib/postgresql/data", "app": "frontend"},
					issues[0].Detail)
			}
		})
	}
}
