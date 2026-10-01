package mcp

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createAppRoutes(api *gin.RouterGroup) {
	api.POST("/projects/p1/prod/apps", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": "a9"}})
	})
}

func newCreateAppWorld(t *testing.T) *writeWorld {
	t.Helper()
	routes := &writeRoutes{}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, createAppRoutes)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

// An app is created empty, as the dashboard's New App does, and the plan says
// what comes next.
func TestCreatingAnApp(t *testing.T) {
	w := newCreateAppWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_create_app",
		map[string]any{"project": "shop", "env": "prod", "name": " Billing ", "tags": []any{"team-a"}})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "plan_redeploy_app")
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "a9")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "POST", sent[0].Method)
		assert.Equal(t, "/projects/p1/prod/apps", sent[0].Path)
		var body map[string]any
		assert.NoError(t, json.Unmarshal([]byte(sent[0].Body), &body))
		assert.Equal(t, "Billing", body["name"], "trimmed")
		assert.Equal(t, "active", body["status"], "as the dashboard creates one")
		assert.Equal(t, []any{"team-a"}, body["tags"])
	}
}

// A name an app of the env has - its name or its key, whatever the case - is
// refused before anything is sent.
func TestCreatingAnAppWithATakenName(t *testing.T) {
	w := newCreateAppWorld(t)
	for _, name := range []string{"API", "api", "worker"} {
		text, isErr := callTool(t, w.session(t, "key1"), "plan_create_app",
			map[string]any{"project": "shop", "env": "prod", "name": name})
		assert.True(t, isErr, name)
		assert.Contains(t, text, "has an app named", name)
	}
	assert.Empty(t, w.routes.writes())
}

func TestDeployImagePrompt(t *testing.T) {
	w := newCreateAppWorld(t)
	args := map[string]string{"project": "shop", "env": "prod", "image": "ghcr.io/acme/api:1.4", "name": "api2"}
	text := promptText(t, w.session(t, "key1"), "deploy_image", args)
	for _, s := range []string{"ghcr.io/acme/api:1.4", "plan_create_app", "kind deployment", "activeMethod image",
		"kind routing", "plan_redeploy_app", "apply it only once I agree"} {
		assert.Contains(t, text, s)
	}
	text = promptText(t, w.session(t, "reader"), "deploy_image", args)
	assert.Contains(t, text, "Change nothing: this key may not")
}
