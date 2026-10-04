package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
)

// The tool's input is the endpoints' body: the files' text as bytes, a port's
// protocol tcp unless said, the apps deployed unless said.
func TestComposeInputIsTheEndpointsBody(t *testing.T) {
	no := false
	req := composeInput{
		Compose: "services: {}", Files: map[string]string{"nginx.conf": "events {}"},
		Project: " Shop ", Images: map[string]string{"app": " me/app:1 "},
		Ports: map[string][]composePortInput{"web": {{Published: 8080, Target: 80, As: "node"}}},
	}.request()

	assert.Equal(t, []byte("events {}"), req.Files["nginx.conf"])
	assert.Equal(t, "Shop", req.Project.Name)
	assert.True(t, req.Deploy)
	assert.Equal(t, "me/app:1", req.Services["app"].Image)
	if assert.Len(t, req.Services["web"].Ports, 1) {
		assert.Equal(t, "tcp", req.Services["web"].Ports[0].Protocol)
		assert.Equal(t, composeservice.PortAsNode, req.Services["web"].Ports[0].As)
	}

	assert.False(t, composeInput{Deploy: &no}.request().Deploy)
}

// Into a project's env, the review's choices of a service named as an app
// there travel with the body.
func TestComposeInputChoosesForAServiceNamedAsAnApp(t *testing.T) {
	req := composeInput{
		Compose: "services: {}", ProjectID: "prj_1", Env: "staging", NewEnv: true,
		Apps: map[string]string{"db": " db2 "}, UseExisting: []string{"cache"},
	}.request()

	assert.True(t, req.Project.NewEnv)
	assert.Equal(t, "staging", req.Project.Env)
	assert.Equal(t, "db2", req.Services["db"].App)
	assert.True(t, req.Services["cache"].UseExisting)
}
