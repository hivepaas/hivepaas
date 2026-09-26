package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
)

// testAuth is a caller: an API key session of user u1.
func testAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User:       &entity.User{ID: "u1", Username: "ada"},
		AuthClaims: &jwtsession.AuthClaims{UserID: "u1", IsAPIKey: true},
	}}
}

// testContext is the context the endpoint gives a tool: the caller, and where
// the MCP request came from.
func testContext(auth *basedto.Auth) context.Context {
	r := httptest.NewRequest(http.MethodPost, "/api/mcp", nil)
	r.RemoteAddr = "203.0.113.7:51000"
	r.Header.Set("User-Agent", "claude-code/2.1")
	r.Header.Set("X-Forwarded-For", "198.51.100.9")
	return withCaller(context.Background(), &caller{auth: auth, keyID: "key1"}, r)
}

// testEngine is a gin engine whose handlers find their caller the way the real
// ones do, through authhandler, with no access check.
func testEngine(routes func(api *gin.RouterGroup, auth *authhandler.Handler)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	auth := authhandler.New(&handler.BaseHandler{}, &sessionuc.UC{})
	routes(engine.Group("/api"), auth)
	return engine
}
