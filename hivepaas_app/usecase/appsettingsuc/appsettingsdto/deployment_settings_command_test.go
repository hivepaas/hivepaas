package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func imageDeploymentReq() *UpdateAppDeploymentSettingsReq {
	return &UpdateAppDeploymentSettingsReq{ProjectID: "p1", ProjectEnvID: "p1:dev", AppID: "a1",
		DeploymentSettingsReq: &DeploymentSettingsReq{ActiveMethod: base.DeploymentMethodImage,
			ImageSource: &DeploymentImageSourceReq{Image: "app:1"}}}
}

func TestDeploymentSettingsCarryTheEntrypoint(t *testing.T) {
	req := imageDeploymentReq()
	req.Entrypoint, req.Command = "/bin/sh -c", `"echo hi && sleep 1"`
	assert.Empty(t, req.Validate())

	settings, err := req.ToEntity()
	assert.NoError(t, err)
	assert.Equal(t, "/bin/sh -c", settings.Entrypoint)
	assert.Equal(t, `"echo hi && sleep 1"`, settings.Command)
}

// A command line that does not split - a quote left open - is refused when it
// is saved, not when a deployment splits it.
func TestDeploymentSettingsRefuseACommandThatDoesNotSplit(t *testing.T) {
	for _, field := range []string{"entrypoint", "command"} {
		t.Run(field, func(t *testing.T) {
			req := imageDeploymentReq()
			if field == "entrypoint" {
				req.Entrypoint = `sh "-c`
			} else {
				req.Command = `echo 'open`
			}
			var paths []string
			for _, inner := range req.Validate().Build(translation.LangEn).InnerErrors {
				paths = append(paths, inner.Path)
			}
			assert.Equal(t, []string{field}, paths)
		})
	}
}
