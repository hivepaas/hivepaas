package server

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/mcp"
)

// Every kind of settings the MCP tools read and change is an endpoint pair of
// the app, and the env link tools' are endpoints of the app too, found here by
// method and path: an endpoint renamed or removed fails
// this test rather than an assistant.
func TestMCPSettingsKindsAreRoutes(t *testing.T) {
	routes := projectRoutes(t)
	const app = "/projects/:projectID/:projectEnv/apps/:appID"
	for _, endpoint := range slices.Concat(mcp.SettingsEndpoints(), mcp.EnvLinkEndpoints()) {
		method, path, _ := strings.Cut(endpoint, " ")
		assert.True(t, routes[method+" "+app+path], "missing route %s %s", method, app+path)
	}
}
