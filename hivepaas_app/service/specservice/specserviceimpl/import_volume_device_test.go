package specserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const projectSettingsPath = "projects/project_a/settings"

// bundleVolume is the project's volume of the fixture, for a test to describe
// differently.
func bundleVolume(bundle *specmodel.ImportBundle) map[string]any {
	volumes, _ := bundle.Projects["project_a"].Settings["volumes"].(map[string]any)
	volume, _ := volumes["default"].(map[string]any)
	return volume
}

// A volume whose data is a directory of the node hands that directory to
// whatever app mounts it, which is what Write on the Cluster module stands for.
func TestPlanRefusesAnImportedVolumeOfTheHostWithoutClusterWrite(t *testing.T) {
	for allowed, wantRefusal := range map[bool]bool{false: true, true: false} {
		svc, bundle := planFixture(t)
		volume := bundleVolume(bundle)
		volume["driver"] = "local"
		volume["driverOpts"] = map[string]any{"type": "none", "o": "bind", "device": "/srv/imported"}

		out, _ := planAskingCluster(t, svc, bundle, allowed)

		settings := node(t, out, projectSettingsPath)
		issues := issuesOf(settings, specmodel.CodeHostMountNotPermitted)
		if !wantRefusal {
			assert.Empty(t, issues)
			assert.Contains(t, settings.Changes, "volumes/default")
			continue
		}
		if assert.Len(t, issues, 1) {
			assert.Equal(t, specmodel.SeveritySkipped, issues[0].Severity)
			assert.Equal(t, "volumes/default", issues[0].Detail[refInSetting])
		}
		assert.NotContains(t, settings.Changes, "volumes/default",
			"the volume is left out; the rest of the scope is imported")
	}
}

// A volume whose directory HivePaaS chooses reaches only the storage HivePaaS
// hands out, and asks nothing.
func TestPlanDoesNotAskAboutAnImportedVolumeOfItsOwnStorage(t *testing.T) {
	svc, bundle := planFixture(t)
	volume := bundleVolume(bundle)
	volume["driver"] = "local"
	volume["driverOpts"] = map[string]any{"size": "100m"}

	out, asked := planAskingCluster(t, svc, bundle, false)

	assert.Zero(t, *asked)
	assert.Empty(t, issuesOf(node(t, out, projectSettingsPath), specmodel.CodeHostMountNotPermitted))
}

// The Docker API is granted through an app's Docker API settings, which bound
// what it may do. A volume carrying the socket would go around them, so a
// bundle asking for one is refused outright.
func TestPlanRefusesABundleWhoseVolumeReachesTheDockerSocket(t *testing.T) {
	for _, device := range []string{"/var/run/docker.sock", "/var/run"} {
		svc, bundle := planFixture(t)
		volume := bundleVolume(bundle)
		volume["driver"] = "local"
		volume["driverOpts"] = map[string]any{"type": "none", "o": "bind", "device": device}

		_, err := svc.planImport(context.Background(), nil, &specservice.ValidateImportReq{
			Scope:               entity.NewObjectScopeGlobal(),
			Options:             specmodel.ImportOptions{Existing: specmodel.ExistingUpdate},
			AllowPrivilegedApps: true,
			Admin:               true,
		}, bundle)

		assert.True(t, errors.Is(err, hperrors.ErrSpecBundleInvalid), "%s: %v", device, err)
	}
}

// A docker mount of a directory holding the socket is the socket, whoever asks.
func TestPlanSkipsAnAppMountingADirectoryHoldingTheDockerSocket(t *testing.T) {
	svc, bundle := planFixture(t)
	backendStorage(bundle).DockerMounts["/run"] = specmodel.Mount{Type: mount.TypeBind, Source: "/var/run"}

	out, _ := planAsAdmin(t, svc, bundle, true)

	backend := node(t, out, backendPath)
	assert.Equal(t, specmodel.ActionSkip, backend.Action)
	if issues := issuesOf(backend, specmodel.CodeHostMountNotPermitted); assert.Len(t, issues, 1) {
		assert.Contains(t, issues[0].Action, "never mounted")
	}
}
