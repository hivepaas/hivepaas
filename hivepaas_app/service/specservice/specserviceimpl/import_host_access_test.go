package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func planAskingCluster(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, allowed bool,
) (*specmodel.ImportPlan, *int) {
	t.Helper()
	asked := 0
	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayWriteCluster: func(context.Context) (bool, error) { asked++; return allowed, nil },
	})
	return out, &asked
}

// The Docker socket mounted into an app is root on its node. An import is the
// one way to ask for a mount of the host, and it takes what capabilities take.
func TestPlanSkipsAnAppMountingTheHostForAnOperatorWhoMayNotWriteTheCluster(t *testing.T) {
	for allowed, wantSkip := range map[bool]bool{false: true, true: false} {
		svc, bundle := planFixture(t)
		backendStorage(bundle).DockerMounts["/var/run/docker.sock"] = specmodel.Mount{
			Type: mount.TypeBind, Source: "/var/run/docker.sock",
		}

		out, asked := planAskingCluster(t, svc, bundle, allowed)

		backend := node(t, out, backendPath)
		assert.Equal(t, 1, *asked)
		if !wantSkip {
			assert.Equal(t, specmodel.ActionUpdate, backend.Action)
			continue
		}
		assert.Equal(t, specmodel.ActionSkip, backend.Action)
		if issues := issuesOf(backend, specmodel.CodeHostMountNotPermitted); assert.Len(t, issues, 1) {
			assert.Equal(t, map[string]any{"mounts": []string{"/var/run/docker.sock"}}, issues[0].Detail,
				"the mount the app already has is not asked about again")
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
