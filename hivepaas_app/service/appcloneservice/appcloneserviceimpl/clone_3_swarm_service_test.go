package appcloneserviceimpl

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func sourceService() *swarm.Service {
	return &swarm.Service{
		ID:   "src",
		Meta: swarm.Meta{Version: swarm.Version{Index: 7}, CreatedAt: time.Now()},
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name:   "src-app",
				Labels: map[string]string{"hivepaas.app.info": "source"},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:    "registry/src:1",
					Env:      []string{"DATABASE_URL=postgres://src"},
					Hostname: "src",
					Labels:   map[string]string{"hivepaas.app.id": "src-app-id"},
				},
			},
			EndpointSpec: &swarm.EndpointSpec{
				Ports: []swarm.PortConfig{{TargetPort: 80, PublishMode: swarm.PortConfigPublishModeHost}},
			},
		},
	}
}

// Each edit here is one cloneSwarmService makes to the clone. With the old
// shallow copy every one of them landed on the source too, and the source's spec
// was later sent back to docker - so cloning a running app with its volume data
// could strip its env, swap its image, drop its host ports and relabel it.
func TestCopyServiceForCloneLeavesTheSourceAlone(t *testing.T) {
	src := sourceService()

	dst, err := copyServiceForClone(src)
	if err != nil {
		t.Fatalf("copyServiceForClone: %v", err)
	}

	dst.Spec.TaskTemplate.ContainerSpec.Env = nil
	dst.Spec.TaskTemplate.ContainerSpec.Image = "init"
	dst.Spec.TaskTemplate.ContainerSpec.Hostname = "clone"
	dst.Spec.TaskTemplate.ContainerSpec.Labels["hivepaas.app.id"] = "clone-app-id"
	dst.Spec.EndpointSpec.Ports = nil
	dst.Spec.Labels["hivepaas.app.info"] = "clone"

	want := sourceService()
	assert.Equal(t, want.Spec.TaskTemplate.ContainerSpec, src.Spec.TaskTemplate.ContainerSpec)
	assert.Equal(t, want.Spec.EndpointSpec, src.Spec.EndpointSpec)
	assert.Equal(t, want.Spec.Labels, src.Spec.Labels)
}

func TestCopyServiceForCloneKeepsTheContent(t *testing.T) {
	src := sourceService()

	dst, err := copyServiceForClone(src)
	if err != nil {
		t.Fatalf("copyServiceForClone: %v", err)
	}

	assert.Equal(t, src.Spec, dst.Spec)
	assert.Equal(t, src.Version, dst.Version)
	assert.NotSame(t, src.Spec.TaskTemplate.ContainerSpec, dst.Spec.TaskTemplate.ContainerSpec)
	assert.NotSame(t, src.Spec.EndpointSpec, dst.Spec.EndpointSpec)
}

// The spec restored at the end of a clone was captured before the service
// existed, with the environment, secrets and config files cleared - the same
// three that applying the copy's configuration writes onto the service between
// then and now. Writing the captured spec back whole undid all of it, so a clone
// asked for its secrets got docker objects attached to nothing and a container
// without the files.
func TestFinalContainerSettingsKeepWhatWasAppliedToTheService(t *testing.T) {
	captured := &swarm.ContainerSpec{
		Image: "registry/src:1", Hostname: "copy",
		Env: nil, Configs: nil, Secrets: nil,
	}
	live := &swarm.ContainerSpec{
		Image: "busybox:latest",
		Env:   []string{"DATABASE_URL=postgres://copy"},
		Configs: []*swarm.ConfigReference{
			{ConfigName: "copy_note.txt", File: &swarm.ConfigReferenceFileTarget{Name: "/data/note.txt"}},
		},
		Secrets: []*swarm.SecretReference{
			{SecretName: "copy_license", File: &swarm.SecretReferenceFileTarget{Name: "/run/secrets/license"}},
		},
	}

	restored := restoreContainerSpec(captured, live)

	assert.Equal(t, "registry/src:1", restored.Image, "the image the clone is meant to run")
	assert.Equal(t, "copy", restored.Hostname)
	assert.Equal(t, live.Env, restored.Env)
	assert.Equal(t, live.Configs, restored.Configs)
	assert.Equal(t, live.Secrets, restored.Secrets)
}

func TestFinalContainerSettingsSurviveAServiceWithNoContainerSpec(t *testing.T) {
	captured := &swarm.ContainerSpec{Image: "registry/src:1"}

	assert.Same(t, captured, restoreContainerSpec(captured, nil))
	assert.Nil(t, restoreContainerSpec(nil, &swarm.ContainerSpec{Env: []string{"A=b"}}))
}
