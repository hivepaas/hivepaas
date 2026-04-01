package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	swaggoGin "github.com/swaggo/gin-swagger"

	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/apphandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/clusterhandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/filehandler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/projecthandler"
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
	authHandler           *authhandler.AuthHandler
	clusterHandler        *clusterhandler.ClusterHandler
	sessionHandler        *sessionhandler.SessionHandler
	userHandler           *userhandler.UserHandler
	projectHandler        *projecthandler.ProjectHandler
	appHandler            *apphandler.AppHandler
	settingHandler        *settinghandler.SettingHandler
	userSettingsHandler   *usersettingshandler.UserSettingsHandler
	systemHandler         *systemhandler.SystemHandler
	systemSettingsHandler *systemsettingshandler.SystemSettingsHandler
	webhookHandler        *webhookhandler.WebhookHandler
	fileHandler           *filehandler.FileHandler
}

func NewHandlerRegistry(
	authHandler *authhandler.AuthHandler,
	clusterHandler *clusterhandler.ClusterHandler,
	sessionHandler *sessionhandler.SessionHandler,
	userHandler *userhandler.UserHandler,
	projectHandler *projecthandler.ProjectHandler,
	appHandler *apphandler.AppHandler,
	settingHandler *settinghandler.SettingHandler,
	userSettingsHandler *usersettingshandler.UserSettingsHandler,
	systemHandler *systemhandler.SystemHandler,
	systemSettingsHandler *systemsettingshandler.SystemSettingsHandler,
	webhookHandler *webhookhandler.WebhookHandler,
	fileHandler *filehandler.FileHandler,
) *HandlerRegistry {
	return &HandlerRegistry{
		authHandler:           authHandler,
		clusterHandler:        clusterHandler,
		sessionHandler:        sessionHandler,
		userHandler:           userHandler,
		projectHandler:        projectHandler,
		appHandler:            appHandler,
		settingHandler:        settingHandler,
		userSettingsHandler:   userSettingsHandler,
		systemHandler:         systemHandler,
		systemSettingsHandler: systemSettingsHandler,
		webhookHandler:        webhookHandler,
		fileHandler:           fileHandler,
	}
}

func (s *HTTPServer) registerRoutes() {
	s.engine.GET("/_/ping", routePing)
	s.engine.NoRoute(routeNotFound)

	// Swagger server
	if !s.config.IsProdEnv() {
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
	if !s.config.IsProdEnv() {
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
}

func routePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func routeNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, "not found")
}
