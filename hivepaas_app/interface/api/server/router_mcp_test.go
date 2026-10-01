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

// The project and env read tools' endpoints are routes, under a project and
// under an env.
func TestMCPProjectEndpointsAreRoutes(t *testing.T) {
	routes := projectRoutes(t)
	projectPaths, envPaths := mcp.ProjectEndpoints()
	for _, path := range projectPaths {
		assert.True(t, routes["GET /projects/:projectID"+path], "missing route GET %s under a project", path)
	}
	for _, path := range envPaths {
		assert.True(t, routes["GET /projects/:projectID/:projectEnv"+path], "missing route GET %s under an env", path)
	}
}

// The project settings kinds are routes, under a project and under an env.
func TestMCPProjectSettingsKindsAreRoutes(t *testing.T) {
	routes := projectRoutes(t)
	projectRoutes, envRoutes := mcp.ProjectSettingsEndpoints()
	for _, endpoint := range projectRoutes {
		method, path, _ := strings.Cut(endpoint, " ")
		assert.True(t, routes[method+" /projects/:projectID"+path], "missing route %s under a project", endpoint)
	}
	for _, endpoint := range envRoutes {
		method, path, _ := strings.Cut(endpoint, " ")
		assert.True(t, routes[method+" /projects/:projectID/:projectEnv"+path], "missing route %s under an env",
			endpoint)
	}
}

// plan_run_sched_job lists an app's or an env's jobs and runs one.
func TestMCPSchedJobRunRoutes(t *testing.T) {
	routes := projectRoutes(t)
	for _, base := range []string{"/projects/:projectID/:projectEnv/apps/:appID", "/projects/:projectID/:projectEnv"} {
		assert.True(t, routes["GET "+base+"/sched-jobs"], "missing route GET %s/sched-jobs", base)
		assert.True(t, routes["POST "+base+"/sched-jobs/:itemID/exec"], "missing route POST %s/sched-jobs/:itemID/exec", base)
	}
}
