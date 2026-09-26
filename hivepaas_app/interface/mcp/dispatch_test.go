package mcp

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
)

func dispatchEngine(served *int) *gin.Engine {
	return testEngine(func(api *gin.RouterGroup, auth *authhandler.Handler) {
		api.GET("/whoami", func(ctx *gin.Context) {
			*served++
			caller, err := auth.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
			if err != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"title": "Unauthorized", "status": 401})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"user": caller.UserID(), "q": ctx.Query("q"),
				"ip": ctx.ClientIP(), "agent": ctx.Request.UserAgent(),
			})
		})
		api.POST("/echo", func(ctx *gin.Context) {
			*served++
			var body map[string]any
			_ = ctx.ShouldBindJSON(&body)
			ctx.JSON(http.StatusOK, body)
		})
		api.GET("/missing", func(ctx *gin.Context) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"title": "Not found", "status": 404, "code": "ERR_NOT_FOUND", "detail": "no app api",
				"cause": "sql: no rows", "stackTrace": "main.go:1",
				"errors": []gin.H{{"path": "app", "message": "unknown"}},
			})
		})
		api.GET("/mcp/loop", func(ctx *gin.Context) { *served++ })
	})
}

func TestDispatchIsAnsweredAsTheCaller(t *testing.T) {
	served := 0
	d := NewDispatcher(dispatchEngine(&served), "/api")
	call := &Call{deps: &Deps{Dispatcher: d}}

	var got map[string]string
	err := call.Get(testContext(testAuth()), "/whoami", url.Values{"q": {"a b"}}, &got)
	assert.NoError(t, err)
	assert.Equal(t, "u1", got["user"])
	assert.Equal(t, "a b", got["q"])
	assert.Equal(t, "claude-code/2.1", got["agent"], "the MCP client, not the backend, is who called")

	var echoed map[string]any
	assert.NoError(t, call.Post(testContext(testAuth()), "/echo", map[string]any{"tail": 10}, &echoed))
	assert.Equal(t, map[string]any{"tail": float64(10)}, echoed)
}

func TestDispatchNeedsACaller(t *testing.T) {
	served := 0
	d := NewDispatcher(dispatchEngine(&served), "/api")
	_, err := d.Do(context.Background(), http.MethodGet, "/whoami", nil, nil)
	assert.ErrorIs(t, err, errNoCaller)
	assert.Zero(t, served)
}

func TestDispatchNeverReachesTheEndpointItself(t *testing.T) {
	served := 0
	d := NewDispatcher(dispatchEngine(&served), "/api")
	for _, path := range []string{"/mcp", "/mcp/loop", "whoami"} {
		_, err := d.Do(testContext(testAuth()), http.MethodGet, path, nil, nil)
		assert.Error(t, err, path)
	}
	assert.Zero(t, served)
}

// An error answer is what the API says, less what is there for debugging.
func TestDispatchReturnsTheAPIsError(t *testing.T) {
	served := 0
	d := NewDispatcher(dispatchEngine(&served), "/api")
	_, err := d.Do(testContext(testAuth()), http.MethodGet, "/missing", nil, nil)
	var apiErr *APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusNotFound, apiErr.Status)
		assert.Equal(t, "ERR_NOT_FOUND", apiErr.Code)
		assert.Equal(t, "404: Not found: no app api; app: unknown", apiErr.Error())
		assert.NotContains(t, apiErr.Error(), "sql")
		assert.NotContains(t, apiErr.Error(), "main.go")
	}
	assert.NotErrorIs(t, err, hperrors.ErrNotFound)
}
