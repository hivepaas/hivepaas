package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//nolint:mnd
func (s *HTTPServer) registerTestRoutes(apiGroup *gin.RouterGroup) *gin.RouterGroup {
	if s.config.IsProdEnv() {
		return nil
	}
	testGroup := apiGroup.Group("/tests")

	testGroup.GET("/req15s", testLongRequest(15*time.Second))
	testGroup.GET("/req30s", testLongRequest(30*time.Second))
	testGroup.GET("/req60s", testLongRequest(60*time.Second))
	testGroup.GET("/req120s", testLongRequest(120*time.Second))
	testGroup.GET("/req180s", testLongRequest(180*time.Second))
	testGroup.GET("/req300s", testLongRequest(300*time.Second))

	return testGroup
}

func testLongRequest(dur time.Duration) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		time.Sleep(dur)
		ctx.String(http.StatusOK, "ok")
	}
}
