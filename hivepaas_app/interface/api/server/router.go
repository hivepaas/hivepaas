package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	swaggoGin "github.com/swaggo/gin-swagger"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appactionhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appcontainerhandler"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectenvsettingshandler"
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
	baseHandler               *handler.BaseHandler
	authHandler               *authhandler.Handler
	appActionHandler          *appactionhandler.Handler
	appContainerHandler       *appcontainerhandler.Handler
	appDeploymentHandler      *appdeploymenthandler.Handler
	appHandler                *apphandler.Handler
	appPreviewHandler         *apppreviewhandler.Handler
	appSettingsHandler        *appsettingshandler.Handler
	clusterHandler            *clusterhandler.Handler
	devHelperHandler          *devhelperhandler.Handler
	fileHandler               *filehandler.Handler
	hivepaasHandler           *hivepaashandler.Handler
	imageHandler              *imagehandler.Handler
	projectEnvHandler         *projectenvhandler.Handler
	projectEnvSettingsHandler *projectenvsettingshandler.Handler
	projectHandler            *projecthandler.Handler
	projectSettingsHandler    *projectsettingshandler.Handler
	sessionHandler            *sessionhandler.Handler
	settingHandler            *settinghandler.Handler
	supportHandler            *supporthandler.Handler
	systemHandler             *systemhandler.Handler
	systemSettingsHandler     *systemsettingshandler.Handler
	traefikHandler            *traefikhandler.Handler
	userHandler               *userhandler.Handler
	userSettingsHandler       *usersettingshandler.Handler
	webhookHandler            *webhookhandler.Handler
}

func NewHandlerRegistry(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	appActionHandler *appactionhandler.Handler,
	appContainerHandler *appcontainerhandler.Handler,
	appDeploymentHandler *appdeploymenthandler.Handler,
	appHandler *apphandler.Handler,
	appPreviewHandler *apppreviewhandler.Handler,
	appSettingsHandler *appsettingshandler.Handler,
	clusterHandler *clusterhandler.Handler,
	devHelperHandler *devhelperhandler.Handler,
	fileHandler *filehandler.Handler,
	hivepaasHandler *hivepaashandler.Handler,
	imageHandler *imagehandler.Handler,
	projectEnvHandler *projectenvhandler.Handler,
	projectEnvSettingsHandler *projectenvsettingshandler.Handler,
	projectHandler *projecthandler.Handler,
	projectSettingsHandler *projectsettingshandler.Handler,
	sessionHandler *sessionhandler.Handler,
	settingHandler *settinghandler.Handler,
	supportHandler *supporthandler.Handler,
	systemHandler *systemhandler.Handler,
	systemSettingsHandler *systemsettingshandler.Handler,
	traefikHandler *traefikhandler.Handler,
	userHandler *userhandler.Handler,
	userSettingsHandler *usersettingshandler.Handler,
	webhookHandler *webhookhandler.Handler,
) *HandlerRegistry {
	return &HandlerRegistry{
		baseHandler:               baseHandler,
		authHandler:               authHandler,
		appActionHandler:          appActionHandler,
		appContainerHandler:       appContainerHandler,
		appDeploymentHandler:      appDeploymentHandler,
		appHandler:                appHandler,
		appPreviewHandler:         appPreviewHandler,
		appSettingsHandler:        appSettingsHandler,
		clusterHandler:            clusterHandler,
		devHelperHandler:          devHelperHandler,
		fileHandler:               fileHandler,
		hivepaasHandler:           hivepaasHandler,
		imageHandler:              imageHandler,
		projectEnvHandler:         projectEnvHandler,
		projectEnvSettingsHandler: projectEnvSettingsHandler,
		projectHandler:            projectHandler,
		projectSettingsHandler:    projectSettingsHandler,
		sessionHandler:            sessionHandler,
		settingHandler:            settingHandler,
		supportHandler:            supportHandler,
		systemHandler:             systemHandler,
		systemSettingsHandler:     systemSettingsHandler,
		traefikHandler:            traefikHandler,
		userHandler:               userHandler,
		userSettingsHandler:       userSettingsHandler,
		webhookHandler:            webhookHandler,
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
	s.engine.Use(StaticServe("/", localFile("./dist-dashboard", true, "")))
	// Serve icons
	s.engine.Use(StaticServe(s.config.HttpPathStaticIcons(),
		embedFile(http.FS(assets.GetIconsFS()), "public, max-age=864000")))
	// Final redirection to redirect any path to `/next=<path>` in case no matching static file found
	s.engine.Use(StaticServeRedirect("/"))

	// INTERNAL ROUTES
	if s.config.IsDevEnv() {
		basicAuthMdlw := gin.BasicAuth(gin.Accounts{
			s.config.Session.BasicAuthUsername: s.config.Session.BasicAuthPassword,
		})
		v1BasicAuth := s.engine.Group(s.config.HTTPServer.BasePath + "/internal")
		v1BasicAuth.Use(basicAuthMdlw)
		s.registerDevRoutes(v1BasicAuth)
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
}

func routePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func routeNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, "not found")
}
