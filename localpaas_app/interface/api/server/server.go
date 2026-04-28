package server

import (
	"context"
	"net/http"
	"time"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/middleware/cors"
	loggermiddleware "github.com/localpaas/localpaas/localpaas_app/interface/api/middleware/logger"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/middleware/recovery"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/middleware/secureheaders"
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
	websocket       *melody.Melody
	handlerRegistry *HandlerRegistry
	logger          logging.Logger
}

// NewHTTPServer Create new LocalPaas app
// @title LocalPaas App
// @version 0.1
// @description LocalPaas App
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
	s.websocket = melody.New()
	s.Server = &http.Server{
		Addr:           s.config.HTTPServer.BindingAddress(),
		ReadTimeout:    180 * time.Second, //nolint:mnd
		WriteTimeout:   180 * time.Second, //nolint:mnd
		MaxHeaderBytes: 1 << 20,           //nolint:mnd
		Handler:        engine,
	}

	// Configures middlewares
	engine.Use(
		recovery.Recovery(s.config),
		loggermiddleware.Logger(s.logger),
		secureheaders.SecureHeaders,
		cors.CORS(s.config),
	)

	if !s.config.IsProdEnv() {
		engine.Use(ginlogger.SetLogger())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.registerRoutes()
}

func (s *HTTPServer) Start() error {
	if s.Server == nil {
		s.init()
	}

	err := s.ListenAndServe()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}
	err := s.Shutdown(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) GetAddress() string {
	return s.Addr
}
