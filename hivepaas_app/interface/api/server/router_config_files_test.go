package server

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projecthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/spechandler"
)

// projectRoutes are the method and path of every route the project group
// registers. The handlers are empty: routing never calls them here.
func projectRoutes(t *testing.T) map[string]bool {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	s := &HTTPServer{handlerRegistry: &HandlerRegistry{
		projectHandler:            &projecthandler.Handler{},
		projectSettingsHandler:    &projectsettingshandler.Handler{},
		projectEnvHandler:         &projectenvhandler.Handler{},
		projectEnvSettingsHandler: &projectenvsettingshandler.Handler{},
		specHandler:               &spechandler.Handler{},
	}}
	s.registerProjectRoutes(engine.Group(""))

	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	return routes
}

func TestProjectAndEnvHaveConfigFileRoutes(t *testing.T) {
	routes := projectRoutes(t)
	for _, base := range []string{"/projects/:projectID/config-files", "/projects/:projectID/:projectEnv/config-files"} {
		for _, want := range []string{
			http.MethodGet + " " + base,
			http.MethodGet + " " + base + "/:itemID",
			http.MethodPost + " " + base,
			http.MethodPut + " " + base + "/:itemID",
			http.MethodPut + " " + base + "/:itemID/status",
			http.MethodDelete + " " + base + "/:itemID",
		} {
			assert.True(t, routes[want], "missing route %s", want)
		}
	}
}
