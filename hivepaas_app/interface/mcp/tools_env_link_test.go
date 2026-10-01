package mcp

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func envLinkSession(t *testing.T) (asked *string, call func(name string, args map[string]any) (string, bool)) {
	t.Helper()
	var target string
	links := func(api *gin.RouterGroup) {
		app := api.Group("/projects/p1/prod/apps/a1/env-vars")
		app.GET("/link-targets", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{
				{"id": "a3", "key": "api-db", "name": "API DB", "category": "database", "engine": "postgres"},
			}})
		})
		app.GET("/link-suggestions", func(ctx *gin.Context) {
			target = ctx.Query("targetAppId")
			ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
				"target": gin.H{"id": target, "key": "api-db", "category": "database", "engine": "postgres"},
				"groups": []gin.H{{"id": "connection-url", "title": "Connection string", "recommended": true,
					"vars": []gin.H{{"key": "DATABASE_URL", "value": "postgres://${api-db.HIVEPAAS_USER}:" +
						"${api-db.HIVEPAAS_PASSWORD_URL_ENCODED}@${api-db.HIVEPAAS_HOST}:${api-db.HIVEPAAS_PORT}/" +
						"${api-db.HIVEPAAS_DATABASE_NAME}?sslmode=${api-db.HIVEPAAS_SSL_MODE}"}}}},
			}})
		})
	}
	w := newMCPWorld(t, (&appRoutes{}).add, links)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return &target, func(name string, args map[string]any) (string, bool) {
		return callTool(t, session, name, args)
	}
}

func TestListEnvLinkTargets(t *testing.T) {
	_, call := envLinkSession(t)
	text, isErr := call("list_env_link_targets", shopProd("api"))
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"api-db"`)
	assert.Contains(t, text, `"postgres"`)
}

// The target is named as any app is - by its key here - and reaches the
// endpoint as the id it takes.
func TestEnvLinkSuggestionsNameTheTargetByKey(t *testing.T) {
	asked, call := envLinkSession(t)
	args := shopProd("api")
	args["target"] = "api-db"
	text, isErr := call("get_env_link_suggestions", args)
	assert.False(t, isErr, text)
	assert.Equal(t, "a3", *asked)
	assert.Contains(t, text, "DATABASE_URL")
	assert.Contains(t, text, "${api-db.HIVEPAAS_SSL_MODE}", "references, not values")

	args["target"] = "nothing-here"
	text, isErr = call("get_env_link_suggestions", args)
	assert.True(t, isErr)
	assert.Contains(t, text, "list_env_link_targets")
}
