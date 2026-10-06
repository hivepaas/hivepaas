package mcp

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func projectSettingsRoutes(api *gin.RouterGroup) {
	api.GET("/projects/p1/env-vars", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"runtimeEnvVars":   []gin.H{{"key": "TZ", "value": "UTC"}},
			"buildtimeEnvVars": []gin.H{}, "updateVer": 4,
		}})
	})
	api.PUT("/projects/p1/env-vars", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"meta": gin.H{}}) })
	api.GET("/projects/p1/prod/env-vars", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"inheritedRuntimeEnvVars": []gin.H{{"key": "TZ", "value": "UTC"}},
			"runtimeEnvVars":          []gin.H{{"key": "LOG_LEVEL", "value": "info"}},
			"buildtimeEnvVars":        []gin.H{}, "inheritedBuildtimeEnvVars": []gin.H{}, "updateVer": 9,
		}})
	})
	api.GET("/projects/p1/domain-settings", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"id": "s1", "rootDomain": "shop.test", "allowedDomains": []string{"*.shop.test"},
		}})
	})
}

func newProjectSettingsWorld(t *testing.T) *writeWorld {
	t.Helper()
	routes := &writeRoutes{}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, projectSettingsRoutes)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

func TestProjectSettingsAreReadWhereTheyAreKept(t *testing.T) {
	w := newProjectSettingsWorld(t)
	session := w.session(t, "key1")

	text, isErr := callTool(t, session, "get_project_settings", map[string]any{"project": "shop", "kind": "env-vars"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"TZ"`)
	assert.NotContains(t, text, `"env"`, "the project's, with no env")

	text, isErr = callTool(t, session, "get_project_settings",
		map[string]any{"project": "shop", "env": "prod", "kind": "env-vars"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"inheritedRuntimeEnvVars"`, "an env's answer shows what it inherits")
	assert.Contains(t, text, `"LOG_LEVEL"`)

	text, isErr = callTool(t, session, "get_project_settings", map[string]any{"project": "shop", "kind": "domain"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"*.shop.test"`)

	text, isErr = callTool(t, session, "get_project_settings",
		map[string]any{"project": "shop", "env": "prod", "kind": "domain"})
	assert.True(t, isErr)
	assert.Contains(t, text, "leave env out")
}

func TestAProjectEnvVarsChangeSendsThePutsRequest(t *testing.T) {
	w := newProjectSettingsWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_update_project_settings", map[string]any{
		"project": "shop", "kind": "env-vars", "changes": map[string]any{
			"runtimeEnvVars": []any{
				map[string]any{"key": "TZ", "value": "America/New_York"},
			},
		},
	})
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	var shown projectSettingsPlan
	assert.NoError(t, json.Unmarshal(plan.Plan, &shown))
	assert.Equal(t, "shop", shown.Project)
	if assert.Len(t, shown.Changes, 1) {
		assert.Equal(t, "runtimeEnvVars", shown.Changes[0].Path)
	}
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "PUT", sent[0].Method)
		assert.Equal(t, "/projects/p1/env-vars", sent[0].Path)
		var body map[string]any
		assert.NoError(t, json.Unmarshal([]byte(sent[0].Body), &body))
		assert.EqualValues(t, 4, body["updateVer"])
		assert.Contains(t, sent[0].Body, "America/New_York")
	}
}

func TestDomainSettingsAreNotChangedThroughTheTools(t *testing.T) {
	w := newProjectSettingsWorld(t)
	text, isErr := callTool(t, w.session(t, "key1"), "plan_update_project_settings", map[string]any{
		"project": "shop", "kind": "domain", "changes": map[string]any{"allowedDomains": []any{"*"}},
	})
	assert.True(t, isErr)
	assert.Contains(t, text, "the dashboard changes them")
	assert.Empty(t, w.routes.writes())
}
