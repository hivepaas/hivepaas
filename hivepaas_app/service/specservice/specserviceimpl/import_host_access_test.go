package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func planAskingCluster(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, allowed bool,
) (*specmodel.ImportPlan, *int) {
	t.Helper()
	asked := 0
	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayWriteCluster:     func(context.Context) (bool, error) { asked++; return allowed, nil },
		AllowPrivilegedApps: true,
	})
	return out, &asked
}

// planAsAdmin plans with the switch on, as an administrator or not, with Write
// on the Cluster module either way, counting how often that is asked.
func planAsAdmin(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, admin bool,
) (*specmodel.ImportPlan, *int) {
	t.Helper()
	asked := 0
	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayWriteCluster:     func(context.Context) (bool, error) { asked++; return true, nil },
		AllowPrivilegedApps: true,
		Admin:               admin,
	})
	return out, &asked
}

var hostDirMount = specmodel.Mount{Type: mount.TypeBind, Source: "/srv/backups"}

// A directory of the node mounted into an app reaches past everything HivePaaS
// keeps apart. An import is the one way to ask for it, and it takes the switch
// and an administrator; Write on the Cluster module is not enough.
func TestPlanSkipsAnAppMountingTheHostForAnOperatorWhoIsNotAnAdministrator(t *testing.T) {
	for admin, wantSkip := range map[bool]bool{false: true, true: false} {
		svc, bundle := planFixture(t)
		backendStorage(bundle).DockerMounts["/backups"] = hostDirMount

		out, asked := planAsAdmin(t, svc, bundle, admin)

		backend := node(t, out, backendPath)
		assert.Zero(t, *asked, "Write on the Cluster module decides nothing here")
		if !wantSkip {
			assert.Equal(t, specmodel.ActionUpdate, backend.Action)
			continue
		}
		assert.Equal(t, specmodel.ActionSkip, backend.Action)
		if issues := issuesOf(backend, specmodel.CodeHostMountNotPermitted); assert.Len(t, issues, 1) {
			assert.Equal(t, map[string]any{"mounts": []string{"/backups"}}, issues[0].Detail,
				"the mount the app already has is not asked about again")
			assert.Contains(t, issues[0].Action, "administrator")
		}
	}
}

// The operator's switch is over what anybody may give an app: with it off, no
// import mounts the host, and nobody is asked.
func TestPlanSkipsAnAppMountingTheHostWhileTheSwitchIsOff(t *testing.T) {
	svc, bundle := planFixture(t)
	backendStorage(bundle).DockerMounts["/backups"] = hostDirMount
	asked := 0

	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayWriteCluster: func(context.Context) (bool, error) { asked++; return true, nil },
		Admin:           true,
	})

	backend := node(t, out, backendPath)
	assert.Zero(t, asked)
	assert.Equal(t, specmodel.ActionSkip, backend.Action)
	if issues := issuesOf(backend, specmodel.CodeHostMountNotPermitted); assert.Len(t, issues, 1) {
		assert.Contains(t, issues[0].Action, "security settings")
	}
}

// An app's socket comes with its access, and the node's socket where host mode
// binds it comes with host mode. Mounting either would hand an app the access
// without anybody granting it, so no import does.
func TestPlanSkipsAnAppMountingASocketItsAccessGives(t *testing.T) {
	for target, m := range map[string]specmodel.Mount{
		"/var/run/hivepaas":    {Type: mount.TypeVolume, Source: dockerapiservice.SocketVolumeName("app_9")},
		"/var/run/docker.sock": {Type: mount.TypeBind, Source: "/var/run/docker.sock"},
	} {
		svc, bundle := planFixture(t)
		backendStorage(bundle).DockerMounts[target] = m

		out, _ := planAsAdmin(t, svc, bundle, true)

		backend := node(t, out, backendPath)
		assert.Equal(t, specmodel.ActionSkip, backend.Action, target)
		if issues := issuesOf(backend, specmodel.CodeHostMountNotPermitted); assert.Len(t, issues, 1, target) {
			assert.Equal(t, map[string]any{"mounts": []string{target}}, issues[0].Detail)
			assert.Contains(t, issues[0].Action, "never mounted", target)
		}
	}
}

// A mount of the host the app already has was allowed when it was made; a
// tmpfs reaches nothing. Neither asks.
func TestPlanDoesNotAskForHostMountsThatChangeNothingOnTheHost(t *testing.T) {
	svc, bundle := planFixture(t)
	storage := backendStorage(bundle)
	storage.DockerMounts["/scratch"] = specmodel.Mount{Type: mount.TypeTmpfs}

	out, asked := planAskingCluster(t, svc, bundle, false)

	assert.Zero(t, *asked)
	assert.Equal(t, specmodel.ActionUpdate, node(t, out, backendPath).Action)
}

// An app is attached only to networks its project can use; the network
// HivePaaS runs its own services on is not one of them.
func TestApplyDropsANetworkTheProjectCannotUse(t *testing.T) {
	svc, bundle := planFixture(t)
	networks := backendDeployment(bundle).Networks
	networks.Attachments = append(networks.Attachments, &specmodel.NetworkAttachment{Name: "hivepaas_net"})

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))

	backend := node(t, resp.Plan, backendPath)
	if issues := issuesOf(backend, specmodel.CodeNetworkNotAvailable); assert.Len(t, issues, 1) {
		assert.Equal(t, specmodel.SeverityFixable, issues[0].Severity)
		assert.Equal(t, map[string]any{"network": "hivepaas_net"}, issues[0].Detail)
	}
	assert.NoError(t, resp.AfterCommit(context.Background(), nil))
	updates := svc.clusterService.(*fakeClusterService).updated["svc_1"]
	if assert.Len(t, updates, 1) {
		var attached []string
		for _, network := range updates[0].TaskTemplate.Networks {
			attached = append(attached, network.Target)
		}
		assert.Equal(t, []string{"project_a_dev_net"}, attached)
	}
}
