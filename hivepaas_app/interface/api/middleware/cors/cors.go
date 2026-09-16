package cors

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	corsCfg := cors.Config{
		AllowOrigins: cfg.HTTPServer.CORSAllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Length", "Origin", "cookie", "access-control-allow-origin",
			"authorization, origin, content-type, accept", "X-CSRF-Token", "Pragma",
			"HIVEPAAS-API-KEY-ID", "HIVEPAAS-API-SECRET-KEY"},
		// A response header the browser may read. Anything not listed here is
		// invisible to a cross-origin caller - which the dashboard is whenever
		// it runs on its own dev server rather than being served by the backend.
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition", "X-HivePaaS-Spec-Report"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, //nolint:mnd
	}
	return cors.New(corsCfg)
}
