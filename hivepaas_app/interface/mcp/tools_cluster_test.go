package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/homeuc/homedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
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
	var out homedto.GetHomeAttentionResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data.Items, 1) {
		assert.Equal(t, "exit 1", out.Data.Items[0].LastError)
		assert.Equal(t, "prod", out.Data.Items[0].Env)
	}
}

// list_tasks is the task list of what its arguments name: everything, an
// env's, or an app's, each asked with the query as the handler reads it.
func TestListTasksIsEveryTaskOrOneEnvs(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_tasks", map[string]any{"status": []string{"failed", "canceled"},
		"pageLimit": 5, "sort": "-createdAt"})
	assert.False(t, isErr, text)
	assert.Equal(t, "pageLimit=5&sort=-createdAt&status=failed%2Ccanceled", routes.asked["/system/tasks"])
	var out taskdto.ListTaskResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data, 1) {
		assert.Equal(t, "build failed", out.Data[0].LastError)
	}

	text, isErr = callTool(t, session, "list_tasks", map[string]any{"project": "shop", "env": "prod"})
	assert.False(t, isErr, text)
	assert.Empty(t, routes.asked["/projects/p1/prod/tasks"])

	text, isErr = callTool(t, session, "list_tasks", map[string]any{"pageLimit": 10001})
	assert.True(t, isErr, text)
	text, isErr = callTool(t, session, "list_tasks", map[string]any{"limit": 5})
	assert.True(t, isErr, "an argument the endpoint does not take is refused: %s", text)
}

func TestGetTaskLogs(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "get_task_logs", map[string]any{"task": "t1", "grep": "error",
		"since": "2026-09-26T08:00:00+07:00", "duration": "2h"})
	assert.False(t, isErr, text)
	assert.Equal(t, "duration=2h&since=2026-09-26T01%3A00%3A00Z&tail=5000&timestamps=true",
		routes.asked["/system/tasks/t1/logs"])
	var out logsAnswer
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []string{"[stderr] error: step 2"}, out.Lines)

	_, isErr = callTool(t, session, "get_task_logs", map[string]any{"task": "t1", "since": "2h"})
	assert.True(t, isErr, "since is a time; a length of time is duration")
}

func TestListNodes(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_nodes", map[string]any{"search": "node"})
	assert.False(t, isErr, text)
	assert.Equal(t, "search=node", routes.asked["/cluster/nodes"])
	var out nodedto.ListNodeResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data, 1) {
		assert.Equal(t, "h1", out.Data[0].Hostname)
		assert.Equal(t, map[string]string{"zone": "a"}, out.Data[0].Labels)
	}
}

func TestTemplates(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "search_templates", map[string]any{"search": "postgres",
		"category": []string{"databases"}})
	assert.False(t, isErr, text)
	assert.Equal(t, "category=databases&search=postgres", routes.asked["/app-templates"])
	var list apptemplatedto.ListAppTemplatesResp
	assert.NoError(t, json.Unmarshal([]byte(text), &list))
	if assert.Len(t, list.Data, 1) && assert.Len(t, list.Data[0].Versions, 1) {
		v := list.Data[0].Versions[0]
		assert.Equal(t, []any{"17", "17.5", true}, []any{v.Name, v.Release, v.Default})
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
		"project": "shop", "env": "prod", "template": "postgres", "name": "db", "imageTag": "17.6",
		"params":           map[string]any{"password": "hunter2-s3cr3t"},
		"dependencyParams": map[string]any{"cache": map[string]any{"password": "hunter3-s3cr3t"}},
	})
	assert.False(t, isErr, text)
	asked := routes.asked["/projects/p1/prod/apps/from-template/preflight"]
	assert.Contains(t, asked, "hunter2-s3cr3t", "the endpoint is asked with the value")
	assert.Contains(t, asked, `"imageTag":"17.6"`)
	assert.NotContains(t, asked, "project", "the env's names are not the request's")
	var out apptemplatedto.PreflightAppFromTemplateResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data.Issues, 1) {
		assert.Equal(t, "ERR_NAME_TAKEN", out.Data.Issues[0].Code)
	}
	if assert.Len(t, out.Data.Storage, 1) {
		assert.True(t, out.Data.Storage[0].IsDatabase)
	}

	if assert.NotEmpty(t, w.audit.entries) {
		detail := w.audit.entries[len(w.audit.entries)-1].Detail
		assert.Contains(t, detail, "password")
		assert.NotContains(t, detail, "s3cr3t")
	}

	_, isErr = callTool(t, session, "preflight_install", map[string]any{"project": "shop", "env": "prod",
		"template": "postgres", "name": "db", "resetStorage": true})
	assert.True(t, isErr, "resetStorage is the dashboard's")
}

func TestSchedules(t *testing.T) {
	_, routes, session := clusterSession(t)
	text, isErr := callTool(t, session, "list_sched_jobs", nil)
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"cronExpr":"0 2 * * *"`)

	text, isErr = callTool(t, session, "explain_schedule", map[string]any{"cronExpr": "0 2 * * *",
		"initialTime": "2026-09-26T19:00:00+07:00", "count": 2})
	assert.False(t, isErr, text)
	var calc map[string]any
	assert.NoError(t, json.Unmarshal([]byte(routes.asked["/settings/sched-jobs/calc-next-runs"]), &calc))
	assert.Equal(t, []any{2.0, "0 2 * * *", "2026-09-26T19:00:00+07:00"},
		[]any{calc["count"], calc["cronExpr"], calc["initialTime"]}, "the endpoint's own request")
	var out schedjobdto.CalcNextRunsResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Len(t, out.Data, 2)

	_, isErr = callTool(t, session, "explain_schedule", map[string]any{"cronExpr": "* * * * *"})
	assert.True(t, isErr, "count is required")
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
