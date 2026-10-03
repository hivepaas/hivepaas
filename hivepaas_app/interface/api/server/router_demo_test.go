package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appactionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appcontainerhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appdeploymenthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apphandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apppreviewhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apptemplatehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/auditloghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/clusterhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/devhelperhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/filehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/hivepaashandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/homehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/imagehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projecthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/sessionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/settinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/spechandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/supporthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/traefikhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/userhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/usersettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/webhookhandler"
)

// allRoutes is every route the API registers, with the handler it calls: the
// handlers are empty, routing never calls them here. MCP's one route and the
// dev helper's are left out; they need a server to be built.
func allRoutes(t *testing.T) gin.RoutesInfo {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	s := &HTTPServer{handlerRegistry: &HandlerRegistry{
		authHandler:               &authhandler.Handler{},
		appActionHandler:          &appactionhandler.Handler{},
		appContainerHandler:       &appcontainerhandler.Handler{},
		appDeploymentHandler:      &appdeploymenthandler.Handler{},
		appHandler:                &apphandler.Handler{},
		appPreviewHandler:         &apppreviewhandler.Handler{},
		appSettingsHandler:        &appsettingshandler.Handler{},
		appTemplateHandler:        &apptemplatehandler.Handler{},
		auditLogHandler:           &auditloghandler.Handler{},
		clusterHandler:            &clusterhandler.Handler{},
		devHelperHandler:          &devhelperhandler.Handler{},
		fileHandler:               &filehandler.Handler{},
		hivepaasHandler:           &hivepaashandler.Handler{},
		homeHandler:               &homehandler.Handler{},
		imageHandler:              &imagehandler.Handler{},
		projectEnvHandler:         &projectenvhandler.Handler{},
		projectEnvSettingsHandler: &projectenvsettingshandler.Handler{},
		projectHandler:            &projecthandler.Handler{},
		projectSettingsHandler:    &projectsettingshandler.Handler{},
		sessionHandler:            &sessionhandler.Handler{},
		settingHandler:            &settinghandler.Handler{},
		specHandler:               &spechandler.Handler{},
		supportHandler:            &supporthandler.Handler{},
		systemHandler:             &systemhandler.Handler{},
		systemSettingsHandler:     &systemsettingshandler.Handler{},
		traefikHandler:            &traefikhandler.Handler{},
		userHandler:               &userhandler.Handler{},
		userSettingsHandler:       &usersettingshandler.Handler{},
		webhookHandler:            &webhookhandler.Handler{},
	}}
	api := engine.Group("/api")
	s.registerSessionRoutes(api)
	s.registerUserRoutes(api)
	s.registerProjectRoutes(api)
	s.registerSettingRoutes(api)
	s.registerSystemRoutes(api)
	s.registerClusterRoutes(api)
	s.registerWebhookRoutes(api)
	s.registerFileRoutes(api)
	s.registerImageRoutes(api)
	s.registerAppTemplateRoutes(api)
	s.registerSupportRoutes(api)
	s.registerHomeRoutes(api)
	s.registerSpecRoutes(api)
	return engine.Routes()
}

// The demo user writes nothing but its session and a schedule's next runs,
// across every route the API has.
func TestTheDemoUserWritesNothingElse(t *testing.T) {
	var allowed []string
	for _, route := range allRoutes(t) {
		if route.Method != http.MethodGet && authhandler.DemoAllows(route.Method, route.Path) {
			allowed = append(allowed, route.Method+" "+route.Path)
		}
	}

	assert.ElementsMatch(t, []string{
		"POST /api/sessions/refresh",
		"DELETE /api/sessions",
		"POST /api/settings/sched-jobs/calc-next-runs",
	}, allowed)
}

// Every route of a handler that acts or hands over what the demo must keep is
// refused to it, at every scope the handler is mounted at - a pattern that no
// longer matches its route would let one through.
func TestTheDemoUserReadsNothingThatActsOrExposes(t *testing.T) {
	refusedHandlers := []string{
		"OpenAppTerminal", "DownloadFileFromContainer",
		"DownloadSecret", "GetSecretDownloadToken",
		"DownloadSSLCertBundle", "DownloadBackupSnapshotFile",
		"BeginGithubAppManifestFlowCreation", "HandleGithubAppManifestFlowProgress",
	}
	seen := map[string]int{}
	for _, route := range allRoutes(t) {
		for _, name := range refusedHandlers {
			if strings.HasSuffix(route.Handler, "."+name+"-fm") {
				seen[name]++
				assert.False(t, authhandler.DemoAllows(route.Method, route.Path),
					"%s %s (%s) is open to the demo user", route.Method, route.Path, name)
			}
		}
	}
	for _, name := range refusedHandlers {
		assert.Positive(t, seen[name], "no route calls %s", name)
	}
}
