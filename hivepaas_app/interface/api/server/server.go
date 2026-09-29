package server

import (
	"context"
	"net/http"
	"time"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/cors"
	loggermiddleware "github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/logger"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/recovery"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/secretguard"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/secureheaders"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/mcp"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

type Server interface {
	Start() error
	Stop(context.Context) error
	GetAddress() string
}

type HTTPServer struct {
	*http.Server
	config          *config.Config
	engine          *gin.Engine
	handlerRegistry *HandlerRegistry
	mcpServices     *mcp.Services
	logger          logging.Logger
}

// NewHTTPServer Create new HivePaaS app
//
// The annotations below are the API's own part of the OpenAPI spec: `make gen-swag`.
// A request authenticates one of two ways, which the two @security lines say: a
// session's access token, or an API key, whose two headers go together (swag's
// "||" joins the schemes of one requirement). The security definitions come
// last: swag reads a definition's lines up to the next definition.
//
// @title HivePaaS API
// @version 0.1
// @description The REST API of HivePaaS, the one its dashboard uses. Every path is under the install's
// @description API base path, `/api`: `GET https://<your HivePaaS domain>/api/projects` lists the projects.
// @description
// @description ## Authentication
// @description
// @description A request authenticates one of two ways:
// @description
// @description - **An API key**, in the headers `HIVEPAAS-API-KEY-ID` and `HIVEPAAS-API-SECRET-KEY`.
// @description A user creates their keys in the dashboard, among their account's settings, or with
// @description `POST /users/current/settings/api-keys`. A key acts as its user, within the limits it
// @description was created with. Scripts and integrations use this.
// @description - **A session's access token**, in the header `Authorization: Bearer <access token>`.
// @description A login such as `POST /auth/login-with-password` gives it, and it expires within
// @description minutes: `POST /sessions/refresh` renews it. The dashboard uses this.
// @description
// @description A request with neither is answered `401`. The logins, the sign-up and password reset
// @description steps, the webhooks and the public images need neither.
// @description
// @description ## Errors
// @description
// @description A request that fails is answered with a 4xx or 5xx status and an `hperrors.ErrorInfo`
// @description body, whose `code` names the error, such as `ERR_NO_SESSION`.
// @contact.name HivePaaS
// @contact.url https://github.com/hivepaas/hivepaas/issues
// @license.name Apache 2.0
// @license.url https://github.com/hivepaas/hivepaas/blob/main/LICENSE
// @BasePath /api
//
// @security APIKeyID || APISecretKey
// @security BearerToken
//
// @tag.name Sessions
// @tag.description Logging in and out, and the session a login opens.
// @tag.name Users
// @tag.description The install's users: invitations, sign-up, profiles, passwords and two-factor authentication.
// @tag.name API keys
// @tag.description The current user's API keys.
// @tag.name Home
// @tag.description What needs attention across the install.
// @tag.name Projects
// @tag.description Projects, which hold the apps.
// @tag.name Project envs
// @tag.description A project's environments. Each holds its own apps and settings.
// @tag.name Project settings
// @tag.description The settings a project holds: credentials, registries, storages, backups and the rest.
// @tag.name Project env settings
// @tag.description The settings a project env holds: credentials, registries, storages, backups and the rest.
// @tag.name Project tasks
// @tag.description The background tasks run in a project.
// @tag.name Project env tasks
// @tag.description The background tasks run in a project env.
// @tag.name Project audit logs
// @tag.description What was done in a project, and by whom.
// @tag.name Project env audit logs
// @tag.description What was done in a project env, and by whom.
// @tag.name Apps
// @tag.description Apps: creating, updating and deleting them, and their logs, terminal and container files.
// @tag.name App templates
// @tag.description The app store: its templates, and creating apps from them.
// @tag.name App settings
// @tag.description An app's settings: env vars, routing, storage, resources, config files, jobs and the rest.
// @tag.name App actions
// @tag.description Deploying, restarting, starting and stopping an app.
// @tag.name App deployments
// @tag.description An app's deployments, their status and logs.
// @tag.name App previews
// @tag.description An app's previews: copies of it deployed for a branch or a pull request.
// @tag.name App tasks
// @tag.description The background tasks run for an app.
// @tag.name Configuration specs
// @tag.description Exporting the install, a project, an env or an app as a configuration spec, and importing one.
// @tag.name Global settings
// @tag.description The settings the whole install holds: credentials, registries, storages, backups and the rest.
// @tag.name Cluster nodes
// @tag.description The Docker Swarm's nodes: joining them, and choosing its managers.
// @tag.name Cluster networks
// @tag.description The Docker Swarm's networks.
// @tag.name Cluster volumes
// @tag.description The Docker Swarm's volumes.
// @tag.name Image builds
// @tag.description The cache the image builds keep.
// @tag.name System
// @tag.description The Get started card, and the database's status.
// @tag.name System settings
// @tag.description The install's own settings: backups, cleanup, logging, the MCP server and the rest.
// @tag.name HivePaaS
// @tag.description HivePaaS itself: its version and updates, and its routing, security and service settings.
// @tag.name Traefik
// @tag.description The Traefik proxy in front of the apps: its config and service settings.
// @tag.name System tasks
// @tag.description The background tasks run for the install.
// @tag.name System audit logs
// @tag.description What was done across the install, and by whom.
// @tag.name System errors
// @tag.description Errors HivePaaS ran into in the background.
// @tag.name Files
// @tag.description Uploading files, and downloading them.
// @tag.name Images
// @tag.description The photos of projects and apps.
// @tag.name Webhooks
// @tag.description The Git webhooks that deploy apps when their repository changes.
// @tag.name Support
// @tag.description Sending feedback to the HivePaaS team.
//
// @securityDefinitions.apikey APIKeyID
// @in header
// @name HIVEPAAS-API-KEY-ID
// @description The ID of an API key. It goes with HIVEPAAS-API-SECRET-KEY.
// @securityDefinitions.apikey APISecretKey
// @in header
// @name HIVEPAAS-API-SECRET-KEY
// @description The secret of an API key. It goes with HIVEPAAS-API-KEY-ID.
// @securityDefinitions.apikey BearerToken
// @in header
// @name Authorization
// @description `Bearer <access token>`, a session's access token from a login.
func NewHTTPServer(
	config *config.Config,
	logger logging.Logger,
	handlerRegistry *HandlerRegistry,
	mcpServices *mcp.Services,
) Server {
	s := &HTTPServer{
		config:          config,
		handlerRegistry: handlerRegistry,
		mcpServices:     mcpServices,
		logger:          logger,
	}
	return s
}

func (s *HTTPServer) init() {
	engine := gin.New()
	s.engine = engine

	applyTrustedProxies(engine, s.config.HTTPServer.TrustedProxies, s.logger)

	s.Server = &http.Server{
		Addr:           s.config.HTTPServer.BindingAddress(),
		ReadTimeout:    180 * time.Second, //nolint:mnd
		WriteTimeout:   180 * time.Second, //nolint:mnd
		MaxHeaderBytes: 1 << 20,           //nolint:mnd
		Handler:        engine,
	}

	// Configures middlewares
	engine.Use(
		recovery.Recovery(s.config, s.handlerRegistry.baseHandler),
		loggermiddleware.Logger(s.logger),
		secureheaders.SecureHeaders,
		cors.CORS(s.config),
	)

	if s.config.IsDevEnv() {
		engine.Use(ginlogger.SetLogger())
		// Development only: it copies every JSON response body, and it panics.
		// Registered after Recovery so the panic is caught and reported.
		engine.Use(secretguard.Guard())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.registerRoutes()
}

// applyTrustedProxies decides whose X-Forwarded-For header is believed.
//
// Gin's own default is to trust 0.0.0.0/0, which means any caller can pick the
// address recorded against them just by sending the header. That is harmless
// while nothing reads the address and wrong the moment something does - the audit
// log does. So nothing is trusted unless it is configured here, and a malformed
// entry falls back to trusting nothing rather than being ignored, which would
// leave gin in the very state this exists to leave.
func applyTrustedProxies(engine *gin.Engine, trustedProxies []string, logger logging.Logger) {
	if err := engine.SetTrustedProxies(trustedProxies); err == nil {
		return
	} else if logger != nil {
		logger.Errorf("invalid http_server.trusted_proxies %v: %v - trusting no proxy",
			trustedProxies, err)
	}
	_ = engine.SetTrustedProxies(nil)
}

func (s *HTTPServer) Start() error {
	if s.Server == nil {
		s.init()
	}

	err := s.ListenAndServe()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}
	err := s.Shutdown(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) GetAddress() string {
	return s.Addr
}
