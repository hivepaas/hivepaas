package cors

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	corsCfg := cors.Config{
		AllowOrigins: cfg.HTTPServer.CORSAllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Length", "Origin", "cookie", "access-control-allow-origin",
			"authorization, origin, content-type, accept", "X-CSRF-Token", "Pragma",
			"LOCALPAAS-API-KEY-ID", "LOCALPAAS-API-SECRET-KEY"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, //nolint:mnd
	}
	return cors.New(corsCfg)
}
