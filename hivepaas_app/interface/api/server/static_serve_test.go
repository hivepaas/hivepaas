package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

// An unknown path is the dashboard's, sent to its index; one under the API's
// base path is the API's, and is left to answer 404 itself.
func TestStaticServeRedirectLeavesTheAPIAlone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := config.Current()
	cfg := &config.Config{}
	cfg.HTTPServer.BasePath = "/api"
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })

	engine := gin.New()
	engine.Use(StaticServeRedirect("/"))
	engine.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })

	for path, want := range map[string]int{
		"/api/no-such-endpoint": http.StatusNotFound,
		"/projects/abc":         http.StatusFound,
		"/apiary":               http.StatusFound,
	} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, want, rec.Code, path)
	}
}
