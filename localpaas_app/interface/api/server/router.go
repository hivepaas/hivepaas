package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	swaggoGin "github.com/swaggo/gin-swagger"

	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appdeploymenthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/apphandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/appsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/clusterhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/devhelperhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/filehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projecthandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projectsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/sessionhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/settinghandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/systemsettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/userhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/usersettingshandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/webhookhandler"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

var (
	defaultPhotoCacheHeader = fmt.Sprintf("public, max-age=%v", int(timeutil.Week.Seconds()))
)

type HandlerRegistry struct {
	authHandler            *authhandler.Handler
	clusterHandler         *clusterhandler.Handler
	sessionHandler         *sessionhandler.Handler
	userHandler            *userhandler.Handler
	projectHandler         *projecthandler.Handler
	projectSettingsHandler *projectsettingshandler.Handler
	appHandler             *apphandler.Handler
	appSettingsHandler     *appsettingshandler.Handler
	appDeploymentHandler   *appdeploymenthandler.Handler
	settingHandler         *settinghandler.Handler
	userSettingsHandler    *usersettingshandler.Handler
	systemHandler          *systemhandler.Handler
	systemSettingsHandler  *systemsettingshandler.Handler
	webhookHandler         *webhookhandler.Handler
	fileHandler            *filehandler.Handler
	devHelperHandler       *devhelperhandler.Handler
}

func NewHandlerRegistry(
	authHandler *authhandler.Handler,
	clusterHandler *clusterhandler.Handler,
	sessionHandler *sessionhandler.Handler,
	userHandler *userhandler.Handler,
	projectHandler *projecthandler.Handler,
	projectSettingsHandler *projectsettingshandler.Handler,
	appHandler *apphandler.Handler,
	appSettingsHandler *appsettingshandler.Handler,
	appDeploymentHandler *appdeploymenthandler.Handler,
	settingHandler *settinghandler.Handler,
	userSettingsHandler *usersettingshandler.Handler,
	systemHandler *systemhandler.Handler,
	systemSettingsHandler *systemsettingshandler.Handler,
	webhookHandler *webhookhandler.Handler,
	fileHandler *filehandler.Handler,
	devHelperHandler *devhelperhandler.Handler,
) *HandlerRegistry {
	return &HandlerRegistry{
		authHandler:            authHandler,
		clusterHandler:         clusterHandler,
		sessionHandler:         sessionHandler,
		userHandler:            userHandler,
		projectHandler:         projectHandler,
		projectSettingsHandler: projectSettingsHandler,
		appHandler:             appHandler,
		appSettingsHandler:     appSettingsHandler,
		appDeploymentHandler:   appDeploymentHandler,
		settingHandler:         settingHandler,
		userSettingsHandler:    userSettingsHandler,
		systemHandler:          systemHandler,
		systemSettingsHandler:  systemSettingsHandler,
		webhookHandler:         webhookHandler,
		fileHandler:            fileHandler,
		devHelperHandler:       devHelperHandler,
	}
}

func (s *HTTPServer) registerRoutes() {
	s.engine.GET("/_/ping", routePing)
	s.engine.NoRoute(routeNotFound)

	// Swagger server
	if s.config.IsDevEnv() {
		s.engine.Use(StaticServe("/docs", localFile("./docs", false, "")))
		s.engine.GET("/swagger/*any", swaggoGin.WrapHandler(swaggoFiles.Handler,
			swaggoGin.URL("/docs/openapi/swagger.json")))
	}

	// STATIC FILES
	s.engine.Use(StaticServe(s.config.HttpPathPhoto(),
		localFile(s.config.DataPathPhoto(), false, defaultPhotoCacheHeader)))
	s.engine.Use(StaticServe(s.config.HttpPathSslLetsEncrypt(),
		localFile(s.config.DataPathSslLetsEncrypt(), false, "")))
	// Serve the static files from the "dist-dashboard" directory at the root URL "/"
	s.engine.Use(StaticServe("/",
		localFile("./dist-dashboard", true, "")))
	// Final redirection to redirect any path to `/next=<path>` in case no matching static file found
	s.engine.Use(StaticServeRedirect("/"))

	// INTERNAL ROUTES
	basicAuthMdlw := gin.BasicAuth(gin.Accounts{
		s.config.Session.BasicAuthUsername: s.config.Session.BasicAuthPassword,
	})
	v1BasicAuth := s.engine.Group(s.config.HTTPServer.BasePath + "/internal")
	v1BasicAuth.Use(basicAuthMdlw)

	// Dev mode
	if s.config.IsDevEnv() {
		v1BasicAuth.POST("/auth/dev-mode-login", s.handlerRegistry.sessionHandler.DevModeLogin)
	}

	// PUBLIC ROUTES
	apiGroup := s.engine.Group(s.config.HTTPServer.BasePath)

	_, _ = s.registerSessionRoutes(apiGroup)
	_, _ = s.registerUserRoutes(apiGroup)
	projectGroup := s.registerProjectRoutes(apiGroup)
	_ = s.registerAppRoutes(projectGroup)
	_ = s.registerSettingRoutes(apiGroup)
	_, _ = s.registerSystemRoutes(apiGroup)
	_ = s.registerClusterRoutes(apiGroup)
	_ = s.registerWebhookRoutes(apiGroup)
	_ = s.registerFileRoutes(apiGroup)
	_ = s.registerDevRoutes(apiGroup)
}

func routePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func routeNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, "not found")
}
