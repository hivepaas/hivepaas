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
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/middleware/secureheaders"
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
	logger          logging.Logger
}

// NewHTTPServer Create new HivePaaS app
// @title HivePaaS App
// @version 0.1
// @description HivePaaS App
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /_
// @securityDefinitions.basic BasicAuth
func NewHTTPServer(
	config *config.Config,
	logger logging.Logger,
	handlerRegistry *HandlerRegistry,
) Server {
	s := &HTTPServer{
		config:          config,
		handlerRegistry: handlerRegistry,
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
