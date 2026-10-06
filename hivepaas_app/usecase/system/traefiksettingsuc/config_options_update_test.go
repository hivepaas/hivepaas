package traefiksettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

// A save with Access Log on writes the option's arguments, the fields it keeps
// among them; an argument of the operator's for one of those is not written, so
// that none drops what HivePaaS counts by - a field kept besides is.
func TestBuildStartupCommandWritesTheAccessLogsArgs(t *testing.T) {
	svc := &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Args: []string{"traefik", "--entrypoints.web.address=:80",
			"--accesslog=true"}},
	}}}
	req := traefiksettingsdto.NewUpdateConfigOptionsReq()
	req.StartupCommand = &traefiksettingsdto.StartupCommandReq{AccessLog: true, Args: []string{
		"--accesslog.fields.names.ServiceName=drop", "--accesslog.fields.defaultmode=keep",
		"--accesslog.fields.names.StartUTC=keep", "--providers.docker=true",
	}}
	assert.NoError(t, req.ModifyRequest())

	want := append([]string{"traefik", "--entrypoints.web.address=:80"}, base.TraefikAccessLogArgs...)
	want = append(want, "--accesslog.fields.names.StartUTC=keep", "--providers.docker=true")
	assert.Equal(t, want, (&UC{}).buildStartupCommand(req, svc))

	req.StartupCommand.AccessLog = false
	assert.Equal(t, []string{"traefik", "--entrypoints.web.address=:80", "--accesslog.fields.names.StartUTC=keep",
		"--providers.docker=true"}, (&UC{}).buildStartupCommand(req, svc), "off: none of the option's")
}
