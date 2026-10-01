package mcp

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func schedRunRoutes(api *gin.RouterGroup) {
	job := func(id, name string) gin.H {
		return gin.H{"id": id, "name": name, "jobType": "container-command", "status": "active",
			"schedule": gin.H{"cronExpr": "0 2 * * *"}}
	}
	api.GET("/projects/p1/prod/apps/a1/sched-jobs", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{job("j1", "nightly-backup"), job("j2", "migrate")}})
	})
	api.GET("/projects/p1/prod/sched-jobs", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{job("j7", "env-cleanup")}})
	})
	exec := func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"task": gin.H{"id": "t42"}}}) }
	api.POST("/projects/p1/prod/apps/a1/sched-jobs/:id/exec", exec)
	api.POST("/projects/p1/prod/sched-jobs/:id/exec", exec)
}

func newSchedRunWorld(t *testing.T) *writeWorld {
	t.Helper()
	routes := &writeRoutes{}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, schedRunRoutes)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

// A job is named as a person names it, and runs once the plan is applied.
func TestRunningAnAppsJobNow(t *testing.T) {
	w := newSchedRunWorld(t)
	session := w.session(t, "key1")
	args := shopProd("api")
	args["job"] = "migrate"
	text, isErr := callTool(t, session, "plan_run_sched_job", args)
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"migrate"`)
	assert.Empty(t, w.routes.writes(), "a plan runs nothing")

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "t42")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "POST", sent[0].Method)
		assert.Equal(t, "/projects/p1/prod/apps/a1/sched-jobs/j2/exec", sent[0].Path)
	}
}

func TestRunningAnEnvsJobNow(t *testing.T) {
	w := newSchedRunWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_run_sched_job",
		map[string]any{"project": "shop", "env": "prod", "job": "j7"})
	assert.False(t, isErr, text)
	_, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/sched-jobs/j7/exec", sent[0].Path)
	}
}

func TestRunningAJobNobodyHas(t *testing.T) {
	w := newSchedRunWorld(t)
	args := shopProd("api")
	args["job"] = "nothing"
	text, isErr := callTool(t, w.session(t, "key1"), "plan_run_sched_job", args)
	assert.True(t, isErr)
	assert.Contains(t, text, "list_sched_jobs")
}
