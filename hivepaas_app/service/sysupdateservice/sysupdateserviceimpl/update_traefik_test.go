package sysupdateserviceimpl

import (
	"context"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

const releaseTraefik = "traefik:v3.7.13@sha256:24841fe2de7304c149343d877d2923b4c8800a38ba015dea9174c23b20e344a0"

// traefikOn is the proxy on image, started with args, its lines not yet marked
// as its own.
func traefikOn(image string, args ...string) *swarm.Service {
	return &swarm.Service{ID: "hivepaas_traefik", Spec: swarm.ServiceSpec{
		Annotations:  swarm.Annotations{Name: "hivepaas_traefik"},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image, Args: args}},
	}}
}

// The access log as an earlier release wrote it: every field.
var earlierAccessLog = []string{"traefik", "--accesslog=true", "--accesslog.format=json",
	"--accesslog.fields.queryparameters.defaultmode=drop"}

// A release that keeps traefik's image still brings its access log to what it
// writes: traefik restarts for it, with rollback armed, on the same image.
func TestTraefikUpdateBringsTheAccessLogToTheReleasesOnTheSameImage(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f, traefikService: fakeTraefik{svc: traefikOn(releaseTraefik, earlierAccessLog...)}}
	data := loggingUpdateData(t, &base.ReleaseInfo{TraefikImage: releaseTraefik})

	assert.NoError(t, s.updateTraefikService(context.Background(), data))

	spec := f.specs["hivepaas_traefik"]
	if assert.NotNil(t, spec, "traefik was updated") {
		cs := spec.TaskTemplate.ContainerSpec
		assert.Equal(t, releaseTraefik, cs.Image)
		assert.Equal(t, append([]string{"traefik"}, base.TraefikAccessLogArgs...), cs.Args)
		assert.Equal(t, "traefik", cs.Labels["hivepaas.component"], "its lines marked as its own")
		if assert.NotNil(t, spec.UpdateConfig) {
			assert.Equal(t, swarm.UpdateFailureActionRollback, spec.UpdateConfig.FailureAction)
		}
	}
}

// One already as the release writes it is not restarted: an update run again
// changes nothing.
func TestTraefikUpdateRestartsNothingAlreadyAsTheReleaseWritesIt(t *testing.T) {
	svc := traefikOn(releaseTraefik, earlierAccessLog...)
	alignTraefik(&svc.Spec)
	f := &fakeDocker{}
	s := &service{dockerManager: f, traefikService: fakeTraefik{svc: svc}}
	data := loggingUpdateData(t, &base.ReleaseInfo{TraefikImage: releaseTraefik})

	assert.NoError(t, s.updateTraefikService(context.Background(), data))

	assert.Empty(t, f.specs)
}

// A newer image takes the release's access log with it, in one restart.
func TestTraefikUpdateMovesTheImageAndTheAccessLogTogether(t *testing.T) {
	f := &fakeDocker{}
	s := &service{dockerManager: f, traefikService: fakeTraefik{svc: traefikOn("traefik:v3.7", earlierAccessLog...)}}
	data := loggingUpdateData(t, &base.ReleaseInfo{TraefikImage: releaseTraefik})

	assert.NoError(t, s.updateTraefikService(context.Background(), data))

	spec := f.specs["hivepaas_traefik"]
	if assert.NotNil(t, spec) {
		assert.Equal(t, releaseTraefik, spec.TaskTemplate.ContainerSpec.Image)
		assert.Equal(t, append([]string{"traefik"}, base.TraefikAccessLogArgs...), spec.TaskTemplate.ContainerSpec.Args)
	}
}

// The plan says so before the update: traefik restarts for its settings, and
// apps are unreachable for the moment it takes. The service is not touched.
func TestPlanSaysTheProxyRestartsForItsSettings(t *testing.T) {
	for _, tc := range []struct {
		name    string
		aligned bool
		change  sysupdateservice.Change
		traffic bool
	}{
		{"as an earlier release wrote it", false, sysupdateservice.ChangeSettings, true},
		{"as this release writes it", true, sysupdateservice.ChangeNone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := traefikOn(releaseTraefik, earlierAccessLog...)
			if tc.aligned {
				alignTraefik(&svc.Spec)
			}
			before := append([]string(nil), svc.Spec.TaskTemplate.ContainerSpec.Args...)
			s := &service{dockerManager: &fakeDocker{}, hpAppService: &fakeHpApp{},
				traefikService:   fakeTraefik{svc: svc},
				systemAppService: &fakeSystemApps{apps: map[string]*entity.App{}}, settingRepo: &fakeSettings{}}

			plan, err := s.PlanUpdate(context.Background(), nil, &base.ReleaseInfo{TraefikImage: releaseTraefik})

			assert.NoError(t, err)
			found := false
			for _, c := range plan.Components {
				if c.Key == base.HivepaasTraefikKey {
					found = true
					assert.Equal(t, tc.change, c.Change)
					assert.Equal(t, tc.traffic, c.InterruptsTraffic)
				}
			}
			assert.True(t, found, "traefik is in the plan")
			assert.Equal(t, before, svc.Spec.TaskTemplate.ContainerSpec.Args, "the plan changes nothing")
		})
	}
}

// logOf is what an update wrote to its task log, a line each.
func logOf(t *testing.T, data *sysUpdateData) string {
	t.Helper()
	frames, err := data.LogStore.GetData(context.Background(), 0)
	assert.NoError(t, err)
	lines := make([]string, 0, len(frames))
	for _, frame := range frames {
		lines = append(lines, frame.Data)
	}
	return strings.Join(lines, "\n")
}

// The log says what the step does: a proxy restarted for its settings is not
// said to be skipped, and one left as it is is.
func TestTraefikUpdateLogSaysWhetherItRestarts(t *testing.T) {
	s := &service{dockerManager: &fakeDocker{},
		traefikService: fakeTraefik{svc: traefikOn(releaseTraefik, earlierAccessLog...)}}
	data := loggingUpdateData(t, &base.ReleaseInfo{TraefikImage: releaseTraefik})

	assert.NoError(t, s.updateTraefikService(context.Background(), data))

	log := logOf(t, data)
	assert.NotContains(t, log, "Skipping traefik")
	assert.Contains(t, log, "traefik: already at traefik:v3.7.13; its settings are brought to this release's")

	aligned := traefikOn(releaseTraefik, earlierAccessLog...)
	alignTraefik(&aligned.Spec)
	s = &service{dockerManager: &fakeDocker{}, traefikService: fakeTraefik{svc: aligned}}
	data = loggingUpdateData(t, &base.ReleaseInfo{TraefikImage: releaseTraefik})

	assert.NoError(t, s.updateTraefikService(context.Background(), data))

	assert.Contains(t, logOf(t, data), "Skipping traefik: already at traefik:v3.7.13")
}
