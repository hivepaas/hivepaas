package mcp

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// previewRoutes answers api's (a1) previews: prepare says what its preview
// settings allow.
func previewRoutes(prepare gin.H) func(api *gin.RouterGroup) {
	return func(api *gin.RouterGroup) {
		app := api.Group("/projects/p1/prod/apps/a1/previews")
		app.POST("/prepare", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"data": prepare}) })
		app.POST("", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"id": "t5"}}) })
		app.GET("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{{"id": "p1", "key": "pr-42", "name": "pr-42",
				"status": "active", "parentId": "a1"}}})
		})
	}
}

func newPreviewWorld(t *testing.T, prepare gin.H) *writeWorld {
	t.Helper()
	routes := &writeRoutes{}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, previewRoutes(prepare))
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

// A preview is planned from what the app's preview settings say - its database
// copies, the secrets it goes without - and sent as the dashboard sends it.
func TestCreatingAPreview(t *testing.T) {
	w := newPreviewWorld(t, gin.H{"enabled": true, "repoURL": "https://github.com/acme/shop",
		"canSkipCloningDbApps": true,
		"withheldSecrets":      []gin.H{{"name": "STRIPE_KEY", "envVars": []string{"PAYMENT_URL"}}}})
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_create_preview", map[string]any{
		"project": "shop", "env": "prod", "app": "api", "repoRef": " pull/42 ", "subdomain": "Review-42",
	})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "STRIPE_KEY")
	assert.Contains(t, text, "PAYMENT_URL")
	assert.Contains(t, text, "copies of the database apps")
	assert.Empty(t, w.routes.writes(), "prepare reads, it changes nothing")

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "t5")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/a1/previews", sent[0].Path)
		body := sentBody(t, sent[0])
		assert.Equal(t, "pull/42", body["repoRef"])
		assert.Equal(t, "review-42", body["customSubdomain"])
		assert.Nil(t, body["cloneDbApps"], "left to the preview settings")
	}
}

func TestAPreviewThatCannotBeMade(t *testing.T) {
	off := newPreviewWorld(t, gin.H{"enabled": false})
	text, isErr := callTool(t, off.session(t, "key1"), "plan_create_preview",
		map[string]any{"project": "shop", "env": "prod", "app": "api", "repoRef": "pull/1"})
	assert.True(t, isErr)
	assert.Contains(t, text, "previews are off for api")

	noDBs := newPreviewWorld(t, gin.H{"enabled": true})
	text, isErr = callTool(t, noDBs.session(t, "key1"), "plan_create_preview",
		map[string]any{"project": "shop", "env": "prod", "app": "api", "repoRef": "pull/1", "cloneDbApps": true})
	assert.True(t, isErr)
	assert.Contains(t, text, "name no database app to clone")
	assert.Empty(t, noDBs.routes.writes())
}

func TestListingAnAppsPreviews(t *testing.T) {
	w := newPreviewWorld(t, gin.H{"enabled": true})
	text, isErr := callTool(t, w.session(t, "reader"), "list_app_previews",
		map[string]any{"project": "shop", "env": "prod", "app": "api"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "pr-42")
}
