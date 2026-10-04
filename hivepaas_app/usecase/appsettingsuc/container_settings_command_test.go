package appsettingsuc

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

// What the container settings screen shows is what a save writes back: a
// shell's script stays one argument, the entrypoint stays the entrypoint, and a
// CMD-SHELL healthcheck stays one string for the shell.
func TestContainerSettingsWriteBackWhatTheyShow(t *testing.T) {
	spec := swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Image:       "app:1",
		Command:     []string{"/bin/sh", "-c"},
		Args:        []string{"npm run migrate && npm start"},
		Healthcheck: &container.HealthConfig{Test: []string{"CMD-SHELL", "pg_isready -U app"}},
	}}}
	shown := appsettingsdto.TransformContainerSettingsBase(&spec)
	assert.Equal(t, "/bin/sh -c", shown.Entrypoint)
	assert.Equal(t, "'npm run migrate && npm start'", shown.Command)
	assert.Equal(t, "pg_isready -U app", shown.Healthcheck.Command)

	data := &updateAppContainerSettingsData{App: &entity.App{ID: "app-1"}, Service: &swarm.Service{Spec: spec}}
	req := &appsettingsdto.UpdateAppContainerSettingsReq{BaseContainerSettings: shown}
	assert.NoError(t, (&UC{}).prepareUpdatingAppContainerSettings(req, data))

	cs := data.Service.Spec.TaskTemplate.ContainerSpec
	assert.Equal(t, []string{"/bin/sh", "-c"}, cs.Command)
	assert.Equal(t, []string{"npm run migrate && npm start"}, cs.Args)
	assert.Equal(t, []string{"CMD-SHELL", "pg_isready -U app"}, cs.Healthcheck.Test)
}

// A healthcheck set to NONE stays NONE: it turns the image's own off.
func TestContainerSettingsKeepAHealthcheckTurnedOff(t *testing.T) {
	data := functionContainerData(nil)
	req := functionContainerReq()
	req.Healthcheck = &appsettingsdto.Healthcheck{Mode: "NONE"}
	assert.NoError(t, (&UC{}).prepareUpdatingAppContainerSettings(req, data))
	assert.Equal(t, []string{"NONE"}, data.Service.Spec.TaskTemplate.ContainerSpec.Healthcheck.Test)
}

// Inherit with no command keeps the image's own test, with the timings given:
// not an empty test of its own, which docker would not run at all.
func TestContainerSettingsInheritTheImagesHealthcheck(t *testing.T) {
	data := functionContainerData(nil)
	req := functionContainerReq()
	req.Healthcheck = &appsettingsdto.Healthcheck{Enabled: true, Interval: timeutil.Duration(time.Minute)}
	assert.NoError(t, (&UC{}).prepareUpdatingAppContainerSettings(req, data))

	check := data.Service.Spec.TaskTemplate.ContainerSpec.Healthcheck
	assert.Empty(t, check.Test)
	assert.Equal(t, time.Minute, check.Interval)

	shown := appsettingsdto.TransformContainerHealthcheck(check)
	assert.Equal(t, docker.HealthcheckModeInherit, shown.Mode)
	assert.True(t, shown.Enabled)
}

// A command line that does not split - a quote left open - is refused before
// it reaches the service.
func TestContainerSettingsRefuseACommandThatDoesNotSplit(t *testing.T) {
	for _, field := range []string{"entrypoint", "command", "healthcheck.command"} {
		t.Run(field, func(t *testing.T) {
			req := functionContainerReq()
			req.ProjectID, req.ProjectEnvID, req.AppID = "p1", "p1:dev", "a1"
			switch field {
			case "entrypoint":
				req.Entrypoint = `sh "-c`
			case "command":
				req.Command = `echo 'open`
			default:
				req.Healthcheck = &appsettingsdto.Healthcheck{Enabled: true, Mode: "CMD", Command: `curl "x`}
			}
			var paths []string
			for _, inner := range req.Validate().Build(translation.LangEn).InnerErrors {
				paths = append(paths, inner.Path)
			}
			assert.Equal(t, []string{field}, paths)
		})
	}
}
