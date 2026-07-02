package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	swaggoGin "github.com/swaggo/gin-swagger"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appactionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appdeploymenthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apphandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/apppreviewhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/clusterhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/devhelperhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/filehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/hivepaashandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/imagehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projecthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/sessionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/settinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/supporthandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/systemsettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/traefikhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/userhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/usersettingshandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/webhookhandler"
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
	appActionHandler       *appactionhandler.Handler
	appPreviewHandler      *apppreviewhandler.Handler
	settingHandler         *settinghandler.Handler
	userSettingsHandler    *usersettingshandler.Handler
	systemHandler          *systemhandler.Handler
	systemSettingsHandler  *systemsettingshandler.Handler
	hivepaasHandler        *hivepaashandler.Handler
	traefikHandler         *traefikhandler.Handler
	webhookHandler         *webhookhandler.Handler
	fileHandler            *filehandler.Handler
	imageHandler           *imagehandler.Handler
	supportHandler         *supporthandler.Handler
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
	appActionHandler *appactionhandler.Handler,
	appPreviewHandler *apppreviewhandler.Handler,
	settingHandler *settinghandler.Handler,
	userSettingsHandler *usersettingshandler.Handler,
	systemHandler *systemhandler.Handler,
	systemSettingsHandler *systemsettingshandler.Handler,
	hivepaasHandler *hivepaashandler.Handler,
	traefikHandler *traefikhandler.Handler,
	webhookHandler *webhookhandler.Handler,
	fileHandler *filehandler.Handler,
	imageHandler *imagehandler.Handler,
	supportHandler *supporthandler.Handler,
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
		appActionHandler:       appActionHandler,
		appPreviewHandler:      appPreviewHandler,
		settingHandler:         settingHandler,
		userSettingsHandler:    userSettingsHandler,
		systemHandler:          systemHandler,
		systemSettingsHandler:  systemSettingsHandler,
		hivepaasHandler:        hivepaasHandler,
		traefikHandler:         traefikHandler,
		webhookHandler:         webhookHandler,
		fileHandler:            fileHandler,
		imageHandler:           imageHandler,
		supportHandler:         supportHandler,
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
	s.engine.Use(StaticServe(s.config.HttpPathSslAcme(),
		localFile(s.config.DataPathSslAcme().AbsPath(), false, "")))
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

	s.registerSessionRoutes(apiGroup)
	s.registerUserRoutes(apiGroup)
	s.registerProjectRoutes(apiGroup)
	s.registerSettingRoutes(apiGroup)
	s.registerSystemRoutes(apiGroup)
	s.registerClusterRoutes(apiGroup)
	s.registerWebhookRoutes(apiGroup)
	s.registerFileRoutes(apiGroup)
	s.registerImageRoutes(apiGroup)
	s.registerSupportRoutes(apiGroup)
	s.registerDevRoutes(apiGroup)
}

func routePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func routeNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, "not found")
}
