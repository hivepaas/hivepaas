package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// clusterRoutes answers the endpoints of the tools beyond apps, and remembers
// what each was asked.
type clusterRoutes struct {
	asked map[string]string // path: raw query, or body
}

func (r *clusterRoutes) remember(ctx *gin.Context) {
	path := strings.TrimPrefix(ctx.Request.URL.Path, "/api")
	if ctx.Request.Method == http.MethodPost {
		var body map[string]any
		_ = ctx.ShouldBindJSON(&body)
		raw, _ := json.Marshal(body)
		r.asked[path] = string(raw)
		return
	}
	r.asked[path] = ctx.Request.URL.RawQuery
}

func (r *clusterRoutes) add(api *gin.RouterGroup) {
	answer := func(data any) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			r.remember(ctx)
			ctx.JSON(http.StatusOK, gin.H{"data": data})
		}
	}
	api.GET("/home/attention", answer(gin.H{"items": []gin.H{{
		"kind": "app-not-running", "severity": "error", "scope": "app", "subject": "api",
		"project": gin.H{"id": "p1", "name": "Shop"}, "env": "prod", "app": gin.H{"id": "a1", "name": "api"},
		"running": 0, "desired": 1, "lastError": "exit 1", "canAct": true,
	}}}))
	tasks := []gin.H{{"id": "t1", "type": "task:app-deploy", "status": "failed", "lastError": "build failed",
		"targetJob": gin.H{"name": "deploy api"}, "scopeApp": gin.H{"key": "api"},
		"createdAt": "2026-09-26T08:00:00Z", "config": gin.H{"anything": "s3cr3t"}}}
	api.GET("/system/tasks", answer(tasks))
	api.GET("/projects/p1/prod/tasks", answer(tasks))
	api.GET("/system/tasks/t1/logs", answer(gin.H{"logs": []logFrame{
		{Type: "out", Data: "step 1"}, {Type: "err", Data: "error: step 2"},
	}}))
	api.GET("/cluster/nodes", answer([]gin.H{{
		"id": "n1", "name": "node-1", "hostname": "h1", "addr": "10.0.0.1", "role": "manager", "isLeader": true,
		"state": "ready", "availability": "active", "labels": gin.H{"zone": "a"},
		"platform":  gin.H{"os": "linux", "architecture": "x86_64"},
		"resources": gin.H{"cpus": 4, "memoryBytes": 8 << 30}, "engineDesc": gin.H{"engineVersion": "28.0"},
	}}))
	api.GET("/app-templates", answer([]gin.H{{
		"name": "postgres", "title": "PostgreSQL", "tagline": "SQL", "categories": []string{"databases/sql"},
		"versions":     []gin.H{{"name": "17", "release": "17.5", "default": true}},
		"dependencies": []gin.H{}, "components": []gin.H{}, "compatible": true,
	}}))
	api.GET("/app-templates/postgres", answer(gin.H{
		"name": "postgres", "title": "PostgreSQL", "tagline": "SQL", "description": "The database.",
		"parameters": []gin.H{{"name": "password", "type": "secret", "generated": true}},
	}))
	api.POST("/projects/p1/prod/apps/from-template/preflight", answer(gin.H{
		"issues":  []gin.H{{"code": "ERR_NAME_TAKEN", "detail": "an app named db exists"}},
		"storage": []gin.H{{"app": "db", "appKey": "db", "isDatabase": true, "volume": gin.H{"name": "data"}}},
	}))
	api.GET("/settings/sched-jobs", answer([]gin.H{{
		"id": "j1", "name": "nightly", "status": "active", "jobType": "command",
		"schedule": gin.H{"cronExpr": "0 2 * * *", "initialTime": "2026-09-01T00:00:00+07:00"},
		"command":  gin.H{"command": "backup --password s3cr3t"},
	}}))
	api.POST("/settings/sched-jobs/calc-next-runs", func(ctx *gin.Context) {
		r.remember(ctx)
		ctx.JSON(http.StatusOK, gin.H{"data": []string{"2026-09-26T19:00:00Z", "2026-09-27T19:00:00Z"}})
	})
}

func clusterSession(t *testing.T) (*mcpWorld, *clusterRoutes, *mcpsdk.ClientSession) {
	t.Helper()
	routes := &clusterRoutes{asked: map[string]string{}}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return w, routes, session
}

func TestListAttention(t *testing.T) {
	_, _, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_attention", nil)
	assert.False(t, isErr, text)
	var out attentionList
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Items, 1) {
		assert.Equal(t, "exit 1", out.Items[0].LastError)
		assert.Equal(t, "prod", out.Items[0].Env)
	}
	assert.NotContains(t, text, "canAct")
}

func TestListTasksIsEveryTaskOrOneEnvs(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_tasks", map[string]any{"status": []string{"failed"}, "limit": 5})
	assert.False(t, isErr, text)
	assert.Equal(t, "pageLimit=5&status=failed", routes.asked["/system/tasks"])
	var out taskList
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Tasks, 1) {
		assert.Equal(t, taskListItem{ID: "t1", Type: "task:app-deploy", Status: "failed", Job: "deploy api",
			App: "api", LastError: "build failed", CreatedAt: time.Date(2026, 9, 26, 8, 0, 0, 0, time.UTC)},
			out.Tasks[0])
	}
	assert.NotContains(t, text, "s3cr3t")

	text, isErr = callTool(t, session, "list_tasks", map[string]any{"project": "shop", "env": "prod"})
	assert.False(t, isErr, text)
	assert.Equal(t, "pageLimit=20", routes.asked["/projects/p1/prod/tasks"])

	text, isErr = callTool(t, session, "list_tasks", map[string]any{"limit": 51})
	assert.True(t, isErr)
	assert.Contains(t, text, "limit is 1 to 50")
}

func TestGetTaskLogs(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "get_task_logs", map[string]any{"task": "t1", "grep": "error"})
	assert.False(t, isErr, text)
	assert.Contains(t, routes.asked["/system/tasks/t1/logs"], "tail=5000")
	var out logsAnswer
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, "t1", out.Task)
	assert.Equal(t, []string{"[stderr] error: step 2"}, out.Lines)
}

func TestListNodes(t *testing.T) {
	_, _, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_nodes", nil)
	assert.False(t, isErr, text)
	var out nodeList
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []nodeItem{{ID: "n1", Name: "node-1", Hostname: "h1", Addr: "10.0.0.1", Role: "manager",
		Leader: true, State: "ready", Availability: "active", CPUs: 4, MemoryBytes: 8 << 30,
		Platform: "linux/x86_64", Engine: "28.0", Labels: map[string]string{"zone": "a"}}}, out.Nodes)
}

func TestTemplates(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "search_templates", map[string]any{"search": "postgres"})
	assert.False(t, isErr, text)
	assert.Equal(t, "pageLimit=50&search=postgres", routes.asked["/app-templates"])
	var list templateList
	assert.NoError(t, json.Unmarshal([]byte(text), &list))
	if assert.Len(t, list.Templates, 1) {
		assert.Equal(t, []templateVersion{{Name: "17", Release: "17.5", Default: true}}, list.Templates[0].Versions)
	}

	text, isErr = callTool(t, session, "get_template", map[string]any{"template": "postgres"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"description":"The database."`)
	assert.Contains(t, text, `"generated":true`)
}

// The parameters of an install may be a password: the audit log keeps their
// names only.
func TestPreflightInstallAuditsNoParameterValue(t *testing.T) {
	w, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "preflight_install", map[string]any{
		"project": "shop", "env": "prod", "template": "postgres", "name": "db",
		"params":           map[string]any{"password": "hunter2-s3cr3t"},
		"dependencyParams": map[string]any{"cache": map[string]any{"password": "hunter3-s3cr3t"}},
	})
	assert.False(t, isErr, text)
	assert.Contains(t, routes.asked["/projects/p1/prod/apps/from-template/preflight"], "hunter2-s3cr3t",
		"the endpoint is asked with the value")
	var out preflightAnswer
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []preflightIssue{{Code: "ERR_NAME_TAKEN", Detail: "an app named db exists"}}, out.Issues)
	assert.Equal(t, []preflightStorage{{App: "db", AppKey: "db", IsDatabase: true, Volume: "data"}}, out.Storage)

	if assert.NotEmpty(t, w.audit.entries) {
		detail := w.audit.entries[len(w.audit.entries)-1].Detail
		assert.Contains(t, detail, "password")
		assert.NotContains(t, detail, "s3cr3t")
	}
}

func TestSchedules(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_sched_jobs", nil)
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"cronExpr":"0 2 * * *"`)
	assert.Contains(t, text, `"initialTime":"2026-09-01T00:00:00+07:00"`)
	assert.NotContains(t, text, "s3cr3t")

	timeNow = func() time.Time { return time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()
	text, isErr = callTool(t, session, "explain_schedule",
		map[string]any{"cronExpr": "0 2 * * *", "timeZone": "Asia/Ho_Chi_Minh", "count": 2})
	assert.False(t, isErr, text)
	assert.JSONEq(t, `{"count":2,"cronExpr":"0 2 * * *","initialTime":"2026-09-26T19:00:00+07:00"}`,
		routes.asked["/settings/sched-jobs/calc-next-runs"])
	var out scheduleRuns
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []string{"2026-09-27T02:00:00+07:00 Sun", "2026-09-28T02:00:00+07:00 Mon"}, out.Runs)

	text, isErr = callTool(t, session, "explain_schedule", map[string]any{"cronExpr": "* * * * *", "timeZone": "Mars"})
	assert.True(t, isErr)
	assert.Contains(t, text, `no time zone "Mars"`)
	_, isErr = callTool(t, session, "explain_schedule", map[string]any{"cronExpr": "* * * * *", "interval": "1h"})
	assert.True(t, isErr)
}

func TestResources(t *testing.T) {
	_, _, session := clusterSession(t)
	ctx := context.Background()
	res, err := session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "hivepaas://templates/postgres"})
	if assert.NoError(t, err) {
		assert.Equal(t, "# PostgreSQL\n\nSQL\n\nThe database.", res.Contents[0].Text)
	}
	_, err = session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "hivepaas://templates/nothing"})
	assert.Error(t, err)

	res, err = session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "hivepaas://docs/docker-api"})
	if assert.NoError(t, err) {
		assert.Contains(t, res.Contents[0].Text, "## Never allowed")
	}
}

func TestPrompts(t *testing.T) {
	_, _, session := clusterSession(t)
	res, err := session.GetPrompt(context.Background(), &mcpsdk.GetPromptParams{Name: "debug_app",
		Arguments: map[string]string{"project": "shop", "env": "prod", "app": "api"}})
	if assert.NoError(t, err) && assert.Len(t, res.Messages, 1) {
		text := res.Messages[0].Content.(*mcpsdk.TextContent).Text
		assert.Contains(t, text, "the app api of shop, env prod")
		assert.Contains(t, text, "get_app_status")
	}
	res, err = session.GetPrompt(context.Background(), &mcpsdk.GetPromptParams{Name: "install_app",
		Arguments: map[string]string{"what": "a database"}})
	if assert.NoError(t, err) {
		text := res.Messages[0].Content.(*mcpsdk.TextContent).Text
		assert.Contains(t, text, "install a database")
		assert.NotContains(t, text, "preflight_install")
	}
}
