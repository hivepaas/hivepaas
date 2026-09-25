package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppEnvVarsHaveLinkRoutes(t *testing.T) {
	routes := projectRoutes(t)
	base := "/projects/:projectID/:projectEnv/apps/:appID/env-vars"
	assert.True(t, routes[http.MethodGet+" "+base+"/link-targets"])
	assert.True(t, routes[http.MethodGet+" "+base+"/link-suggestions"])
}
