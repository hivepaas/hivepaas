package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/logging"
)

// Every deployment sets the container's arguments from the command string, split
// the way a shell would. What vlagent is given has JSON and spaces in it; it has
// to arrive as it was written.
func TestCommandLineSurvivesTheSplit(t *testing.T) {
	args := []string{
		"-remoteWrite.url=http://victoria-logs:9428/internal/insert",
		`-fileCollector.extraFields={"app":"a b","quote":"it's"}`,
		"-remoteWrite.headers=X-Scope: one two^^X-Other: $HOME",
		"-remoteWrite.bearerTokenFile=",
	}

	split, err := executil.CmdSplit(commandLine(args))

	assert.NoError(t, err)
	assert.Equal(t, args, split)
}

// The document goes through the same check a template's does, credentials
// included, so the stack cannot ask for something an app cannot be.
func TestBuildAppDocIsBuildable(t *testing.T) {
	rt := &logging.RuntimeSpec{
		Image:     "victoriametrics/vlagent:v1.52.0",
		Args:      []string{"-a=1"},
		Mounts:    []logging.Mount{{Volume: "vol-1", Target: "/data"}, {Source: "/host", Target: "/host"}},
		Secrets:   []*logging.Secret{{Key: "REMOTE_WRITE_1_PASSWORD", Path: "/run/secrets/p", Value: "v"}},
		Resources: logging.Resources{MemoryLimit: (512 * unit.MB).Bytes()},
	}

	doc, err := buildAppDoc(rt, "vlagent")

	assert.NoError(t, err)
	assert.Len(t, doc.Deployment.Storage.Mounts, 1, "a host path is not a document's to name")
	assert.Contains(t, doc.Settings, "secrets")
}

// A document has no say over the mode, host paths, the stack's networks or the
// log driver; customizeSpec writes them.
func TestCustomizeSpecWritesWhatADocumentCannot(t *testing.T) {
	spec := &swarm.ServiceSpec{
		Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{}},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
	}
	rt := &logging.RuntimeSpec{
		PerNode: true,
		Mounts:  []logging.Mount{{Source: "/var/lib/docker/containers", Target: "/c", ReadOnly: true}},
	}

	assert.NoError(t, customizeSpec(rt, []string{"hivepaas_local_net"}, "alias")(spec))

	assert.NotNil(t, spec.Mode.Global)
	assert.Nil(t, spec.Mode.Replicated)
	assert.Len(t, spec.TaskTemplate.ContainerSpec.Mounts, 1)
	assert.Equal(t, &swarm.Driver{Name: logDriverLocal}, spec.TaskTemplate.LogDriver)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "hivepaas_local_net", Aliases: []string{"alias"}}},
		spec.TaskTemplate.Networks)
}

// A volume reaches an app only when it is shared with apps.
func TestValidateRefusesAVolumeNotSharedWithApps(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	settingsOf(s).volume.Inheritable = false

	err := s.Validate(context.Background(), nil, enabledConfig(), nil)

	assert.ErrorIs(t, err, hperrors.ErrLoggingSettingsInvalid)
}

// A running backend keeps its volume: nothing copies the logs to another one.
func TestValidateRefusesToMoveARunningBackend(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	next := enabledConfig()
	next.Backend.VictoriaLogs.Volume = entity.ObjectID{ID: "vol-2"}
	err = s.Validate(context.Background(), nil, next, enabledConfig())

	assert.ErrorIs(t, err, hperrors.ErrLoggingVolumeImmutable)
}

// The settings screen reads the apps, and the backend's limits off its service.
func TestStatusReadsTheAppsAndTheirLimits(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	status, err := s.Status(context.Background(), nil, storedSetting(t, enabledConfig()))
	assert.NoError(t, err)
	assert.Nil(t, status.Backend)
	assert.Equal(t, defaultBackendResources(), status.BackendResources, "what a new backend would get")

	_, err = apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)
	backend := appsOf(s).docker.inspected["svc-"+backendAppKey]
	setResources(&backend.Spec, logging.Resources{MemoryLimit: (3 * unit.GB).Bytes()})
	backend.ServiceStatus = &swarm.ServiceStatus{RunningTasks: 1, DesiredTasks: 1}
	appsOf(s).docker.inspected["svc-"+backendAppKey] = backend

	status, err = s.Status(context.Background(), nil, storedSetting(t, enabledConfig()))

	assert.NoError(t, err)
	assert.True(t, status.Enabled)
	assert.True(t, status.Backend.Ready())
	assert.Equal(t, "app-"+backendAppKey, status.Backend.AppID)
	assert.Equal(t, (3 * unit.GB).Bytes(), status.BackendResources.MemoryLimit)
	assert.False(t, status.Collector.Ready(), "no task is running yet")
}
