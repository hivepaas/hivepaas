package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A function is created among its environment's apps, beside an app from a
// template, without taking the place of an app's id.
func TestAFunctionIsCreatedAmongTheEnvsApps(t *testing.T) {
	routes := projectRoutes(t)

	assert.True(t, routes[http.MethodPost+" /projects/:projectID/:projectEnv/apps/function"])
	assert.True(t, routes[http.MethodPost+" /projects/:projectID/:projectEnv/apps/from-template"])
	assert.True(t, routes[http.MethodPut+" /projects/:projectID/:projectEnv/apps/:appID"])
}

// A function's test run is a call on the app.
func TestAFunctionIsTestRunOnItsApp(t *testing.T) {
	routes := projectRoutes(t)

	assert.True(t, routes[http.MethodPost+" /projects/:projectID/:projectEnv/apps/:appID/function/test-run"])
}
