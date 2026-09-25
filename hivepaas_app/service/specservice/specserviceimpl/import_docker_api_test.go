package specserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type fakeImportDockerAPI struct {
	dockerapiservice.Service
	calls []string
}

func (f *fakeImportDockerAPI) EnsureNetwork(_ context.Context, appID string) (string, error) {
	f.calls = append(f.calls, "network "+appID)
	return "net-" + appID, nil
}

func (f *fakeImportDockerAPI) SyncAgents(context.Context) error {
	f.calls = append(f.calls, "sync")
	return nil
}

// ApplyToService attaches as the real one does for an app with access.
func (f *fakeImportDockerAPI) ApplyToService(
	_ context.Context, _ database.IDB, appID string, spec *swarm.ServiceSpec,
) error {
	f.calls = append(f.calls, "apply "+appID)
	dockerapiservice.Attach(spec, appID, "net-"+appID)
	return nil
}

func dockerAPIBody(images ...any) map[string]any {
	return map[string]any{"images": images, "limits": map[string]any{"containers": 3}}
}

// Giving an app the Docker API takes what giving it from a template takes.
func TestPlanSkipsAnAppGivenTheDockerAPIByAnOperatorWhoMayNotWriteTheCluster(t *testing.T) {
	for allowed, wantSkip := range map[bool]bool{false: true, true: false} {
		svc, bundle := planFixture(t)
		backendSettings(bundle)["dockerApi"] = dockerAPIBody("autobase/automation")

		out, asked := planAskingCluster(t, svc, bundle, allowed)

		backend := node(t, out, backendPath)
		assert.Equal(t, 1, *asked)
		if !wantSkip {
			assert.Equal(t, specmodel.ActionUpdate, backend.Action)
			assert.Contains(t, backend.Changes, "settings.dockerApi")
			assert.True(t, backend.Restart, "the service gains its socket")
			continue
		}
		assert.Equal(t, specmodel.ActionSkip, backend.Action)
		assert.Len(t, issuesOf(backend, specmodel.CodeDockerAPINotPermitted), 1)
	}
}

// The node's own socket takes the switch and an administrator, and nothing
// else: Write on the Cluster module is not asked.
func TestPlanGivesHostModeOnlyToAnAdministratorWithTheSwitchOn(t *testing.T) {
	for name, tc := range map[string]struct {
		on, admin bool
		missing   string
	}{
		"the switch off":       {admin: true, missing: "System → HivePaaS → Security"},
		"not an administrator": {on: true, missing: "administrator"},
		"both":                 {on: true, admin: true},
	} {
		svc, bundle := planFixture(t)
		backendSettings(bundle)["dockerApi"] = map[string]any{"mode": "host"}
		asked := 0

		out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
			MayWriteCluster:     func(context.Context) (bool, error) { asked++; return true, nil },
			AllowPrivilegedApps: tc.on,
			Admin:               tc.admin,
		})

		backend := node(t, out, backendPath)
		assert.Zero(t, asked, name)
		if tc.missing == "" {
			assert.Equal(t, specmodel.ActionUpdate, backend.Action, name)
			continue
		}
		assert.Equal(t, specmodel.ActionSkip, backend.Action, name)
		if issues := issuesOf(backend, specmodel.CodeDockerSocketNotPermitted); assert.Len(t, issues, 1, name) {
			assert.Contains(t, issues[0].Action, tc.missing, name)
		}
	}
}

// A new app in host mode is provisioned with the node's socket, and no network
// of its own: host mode has no use for one.
func TestApplyGivesANewAppInHostModeTheNodesSocket(t *testing.T) {
	svc, bundle := planFixture(t)
	dockerAPI := &fakeImportDockerAPI{}
	svc.dockerAPIService = dockerAPI
	addWorker(bundle, map[string]any{"dockerApi": map[string]any{"mode": "host"}})
	req := &specservice.ApplyImportReq{OperatorID: "u_operator", AcceptIssues: true}
	req.AllowPrivilegedApps, req.Admin = true, true
	req.PlanHash = planWith(t, svc, bundle, &req.ValidateImportReq).PlanHash

	apply(t, svc, bundle, req)

	provision := svc.appProvisionService.(*fakeProvisionService)
	if !assert.Len(t, provision.reqs, 1) {
		return
	}
	spec := provision.specs[provision.reqs[0].AppID]
	assert.Equal(t, []mount.Mount{dockerapiservice.HostSocketMount()},
		spec.TaskTemplate.ContainerSpec.Mounts[len(spec.TaskTemplate.ContainerSpec.Mounts)-1:])
	assert.Empty(t, spec.TaskTemplate.Networks)
	assert.Empty(t, dockerAPI.calls)
}

// A block the screen or a template would refuse is refused here too, rather
// than written as a policy the proxy cannot make sense of.
func TestPlanRefusesADockerAPIBlockWrongInItself(t *testing.T) {
	svc, bundle := planFixture(t)
	backendSettings(bundle)["dockerApi"] = dockerAPIBody()

	_, err := svc.planImport(context.Background(), nil, &specservice.ValidateImportReq{
		Scope: entity.NewObjectScopeGlobal(), Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate},
	}, bundle)

	assert.True(t, errors.Is(err, hperrors.ErrSpecBundleInvalid), "%v", err)
}

// Export leaves the socket and the network out: they are named after this
// installation's app. A new app gets its own while it is provisioned.
func TestApplyGivesANewAppTheSocketAndNetworkOfItsAccess(t *testing.T) {
	svc, bundle := planFixture(t)
	dockerAPI := &fakeImportDockerAPI{}
	svc.dockerAPIService = dockerAPI
	addWorker(bundle, map[string]any{"dockerApi": dockerAPIBody("*")})

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))

	provision := svc.appProvisionService.(*fakeProvisionService)
	if !assert.Len(t, provision.reqs, 1) {
		return
	}
	appID := provision.reqs[0].AppID
	spec := provision.specs[appID]
	assert.Contains(t, spec.TaskTemplate.ContainerSpec.Mounts, dockerapiservice.SocketMount(appID))
	assert.Contains(t, spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: "net-" + appID})

	assert.NoError(t, resp.AfterCommit(context.Background(), nil))
	assert.Equal(t, []string{"network " + appID, "sync"}, dockerAPI.calls)
}

// An existing app given access gets its socket in the one update phase two
// makes, once every agent serves it: a task that starts before its socket
// exists finds nothing, and many apps give up on that.
func TestPhaseTwoGivesAnUpdatedAppItsSocketOnceTheAgentsServeIt(t *testing.T) {
	svc, bundle := planFixture(t)
	dockerAPI := &fakeImportDockerAPI{}
	svc.dockerAPIService = dockerAPI
	backendSettings(bundle)["dockerApi"] = dockerAPIBody("*")

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))
	assert.NoError(t, resp.AfterCommit(context.Background(), nil))

	updates := svc.clusterService.(*fakeClusterService).updated["svc_1"]
	if assert.Len(t, updates, 1) {
		assert.Contains(t, updates[0].TaskTemplate.ContainerSpec.Mounts, dockerapiservice.SocketMount("app_1"))
	}
	assert.Equal(t, []string{"sync", "apply app_1"}, dockerAPI.calls)
}

// Rebuilding an app's deployment blocks writes its mounts and networks afresh;
// what its access gives it is put back in the same update.
func TestPhaseTwoKeepsTheSocketOfAnAppWhoseDeploymentChanged(t *testing.T) {
	svc, bundle := planFixture(t)
	dockerAPI := &fakeImportDockerAPI{}
	svc.dockerAPIService = dockerAPI
	backendDeployment(bundle).Resources.Limits = &specmodel.ResourceLimits{CPUs: 2}

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))
	assert.NoError(t, resp.AfterCommit(context.Background(), nil))

	updates := svc.clusterService.(*fakeClusterService).updated["svc_1"]
	if assert.Len(t, updates, 1) {
		assert.Contains(t, updates[0].TaskTemplate.ContainerSpec.Mounts, dockerapiservice.SocketMount("app_1"))
	}
	assert.Equal(t, []string{"apply app_1"}, dockerAPI.calls, "nothing to sync: no access was written")
}
