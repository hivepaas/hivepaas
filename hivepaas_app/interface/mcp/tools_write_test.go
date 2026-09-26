package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// sentRequest is a request that would change something.
type sentRequest struct {
	Method, Path, Body string
}

// writeRoutes answers what the write tools read and send, and remembers every
// request that is not a read.
type writeRoutes struct {
	mu     sync.Mutex
	sent   []sentRequest
	source gin.H
	// preflight is what the install preflight answers.
	preflight gin.H
	// importPlan is what the env's import validate answers; validated is the
	// app document of the last bundle it was sent; planChanged makes apply
	// refuse, as the import does when the target moved.
	importPlan  gin.H
	validated   string
	planChanged bool
}

func (r *writeRoutes) writes() []sentRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]sentRequest(nil), r.sent...)
}

func (r *writeRoutes) add(api *gin.RouterGroup) {
	api.Use(func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet && !strings.HasSuffix(ctx.Request.URL.Path, "/validate") &&
			!strings.HasSuffix(ctx.Request.URL.Path, "/preflight") &&
			!strings.HasSuffix(ctx.Request.URL.Path, "/spec/export") &&
			!strings.HasSuffix(ctx.Request.URL.Path, "/calc-next-runs") {
			body, _ := io.ReadAll(ctx.Request.Body)
			r.mu.Lock()
			r.sent = append(r.sent, sentRequest{ctx.Request.Method,
				strings.TrimPrefix(ctx.Request.URL.Path, "/api"), string(body)})
			r.mu.Unlock()
			ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))
		}
		ctx.Next()
	})
	api.POST("/projects/p1/prod/apps/from-template/preflight", func(ctx *gin.Context) {
		r.mu.Lock()
		defer r.mu.Unlock()
		ctx.JSON(http.StatusOK, gin.H{"data": r.preflight})
	})
	api.POST("/projects/p1/prod/apps/from-template", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"data": gin.H{
			"app": gin.H{"id": "n1"}, "deployment": gin.H{"id": "d1"},
			"dependencies": []gin.H{{"name": "db", "app": gin.H{"id": "n0"}, "deployment": gin.H{"id": "d0"}}},
		}})
	})
	api.POST("/projects/p1/prod/spec/import/validate", func(ctx *gin.Context) {
		var body struct {
			Bundle    []byte `json:"bundle"`
			Selection struct {
				Include []string `json:"include"`
			} `json:"selection"`
		}
		_ = ctx.ShouldBindJSON(&body)
		doc, _ := appDocument(body.Bundle, "api")
		r.mu.Lock()
		defer r.mu.Unlock()
		r.validated = strings.Join(body.Selection.Include, ",") + "\n" + doc
		ctx.JSON(http.StatusOK, gin.H{"data": r.importPlan})
	})
	api.POST("/projects/p1/prod/spec/import/apply", func(ctx *gin.Context) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.planChanged {
			ctx.JSON(http.StatusConflict, gin.H{"status": 409, "title": "Conflict",
				"code": "ERR_SPEC_IMPORT_PLAN_CHANGED"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"plan": gin.H{"nodes": []gin.H{
			{"path": "projects/shop/envs/prod/apps/api", "action": "update", "outcome": "updated"},
		}}, "deployments": []gin.H{}}})
	})
	api.POST("/settings/sched-jobs/calc-next-runs", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": []string{"2026-09-26T19:00:00Z", "2026-09-27T19:00:00Z"}})
	})
	api.POST("/projects/p1/prod/apps/a1/sched-jobs", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": "j9"}})
	})
	app := api.Group("/projects/p1/prod/apps/a1")
	app.GET("/deployment-settings", func(ctx *gin.Context) {
		r.mu.Lock()
		defer r.mu.Unlock()
		ctx.JSON(http.StatusOK, gin.H{"data": r.source})
	})
	app.POST("/restart", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"meta": gin.H{}}) })
	app.POST("/deploy", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"deploymentId": "d9"}})
	})
}

type writeWorld struct {
	*mcpWorld
	routes *writeRoutes
}

func newWriteWorld(t *testing.T) *writeWorld {
	t.Helper()
	routes := &writeRoutes{source: gin.H{"activeMethod": "image", "imageSource": gin.H{"image": "shop/api:1.25"}},
		preflight: gin.H{"issues": []gin.H{}, "apps": []gin.H{
			{"name": "wiki-db", "key": "wiki-db", "kind": "dependency", "role": "db", "template": "postgres",
				"image": "postgres:17"},
			{"name": "wiki", "key": "wiki", "kind": "app", "template": "wikijs", "image": "wikijs:2"},
		}}}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

func (w *writeWorld) session(t *testing.T, key string) *mcpsdk.ClientSession {
	t.Helper()
	session, err := w.connect(t, key)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return session
}

func toolNames(t *testing.T, session *mcpsdk.ClientSession) []string {
	t.Helper()
	listed, err := session.ListTools(context.Background(), nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	names := make([]string, 0, len(listed.Tools))
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	return names
}

func planOf(t *testing.T, text string) planResult[json.RawMessage] {
	t.Helper()
	var out planResult[json.RawMessage]
	if !assert.NoError(t, json.Unmarshal([]byte(text), &out), text) {
		t.FailNow()
	}
	return out
}

// Changing anything takes the setting, a key that may change things, and the
// endpoint's own permission. Without the first two a client is not even listed
// the tools, and a plan made while they held is not applied once they do not.
func TestWritesNeedAllThreeSwitches(t *testing.T) {
	w := newWriteWorld(t)

	assert.Contains(t, toolNames(t, w.session(t, "key1")), "apply_plan")
	assert.NotContains(t, toolNames(t, w.session(t, "reader")), "apply_plan", "a key that may only read")
	assert.NotContains(t, toolNames(t, w.session(t, "reader")), "plan_restart_app")

	w.sw.allowWrite = false
	assert.NotContains(t, toolNames(t, w.session(t, "key1")), "apply_plan", "changes not allowed")
	assert.NotContains(t, toolNames(t, w.session(t, "key1")), "plan_restart_app")

	w.sw.allowWrite = true
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_restart_app", shopProd("api"))
	assert.False(t, isErr, text)
	token := planOf(t, text).PlanToken

	w.sw.allowWrite = false
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: "apply_plan",
		Arguments: map[string]any{"planToken": token}})
	assert.True(t, err != nil || res.IsError, "an earlier plan is not applied once changes are off")
	assert.Empty(t, w.routes.writes())

	w.sw.allowWrite = true
	text, isErr = callTool(t, w.session(t, "key2"), "apply_plan", map[string]any{"planToken": token})
	assert.True(t, isErr)
	assert.Contains(t, text, "no such plan", "another key of the same user")
	assert.Empty(t, w.routes.writes())
}

func TestAPlanChangesNothingAndApplySendsExactlyIt(t *testing.T) {
	w := newWriteWorld(t)
	session := w.session(t, "key1")

	text, isErr := callTool(t, session, "plan_restart_app", shopProd("api"))
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	assert.True(t, strings.HasPrefix(plan.PlanToken, "mcpp_"))
	assert.Contains(t, plan.Next, "only once they have agreed")
	assert.Contains(t, string(plan.Plan), `"tasksNow"`)
	assert.Empty(t, w.routes.writes(), "a plan sends nothing")

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"summary":"restart api in shop/prod"`)
	assert.Equal(t, []sentRequest{{"POST", "/projects/p1/prod/apps/a1/restart", "{}"}}, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.True(t, isErr)
	assert.Contains(t, text, "no such plan", "a plan is used once")
	assert.Len(t, w.routes.writes(), 1)

	// The plan's entry and the apply's name the same plan; neither holds the token.
	var planned, applied string
	for _, entry := range w.audit.entries {
		switch {
		case entry.ResName == "plan_restart_app":
			planned = entry.Detail
		case entry.ResName == "apply_plan: plan_restart_app":
			applied = entry.Detail
		}
		assert.NotContains(t, entry.Detail, plan.PlanToken)
	}
	planID, _, _ := strings.Cut(strings.TrimPrefix(plan.PlanToken, "mcpp_"), ".")
	assert.Contains(t, planned, planID)
	assert.Contains(t, applied, planID)
	assert.Contains(t, applied, "plan_restart_app")
}

func TestRedeployAnotherTag(t *testing.T) {
	w := newWriteWorld(t)
	session := w.session(t, "key1")
	args := shopProd("api")
	args["imageTag"] = "1.27"

	text, isErr := callTool(t, session, "plan_redeploy_app", args)
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	var shown redeployPlan
	assert.NoError(t, json.Unmarshal(plan.Plan, &shown))
	assert.Equal(t, "shop/api:1.25", shown.Now.Image)
	assert.Equal(t, "shop/api:1.27", shown.After.Image)

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"deploymentId":"d9"`)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/a1/deploy", sent[0].Path)
		assert.JSONEq(t, `{"activeMethod":"image","imageSource":{"imageTag":"1.27"}}`, sent[0].Body)
	}
}

// The person agreed to deploy what they were shown: a source changed since is
// not deployed.
func TestRedeployRefusesASourceThatMoved(t *testing.T) {
	w := newWriteWorld(t)
	session := w.session(t, "key1")
	text, _ := callTool(t, session, "plan_redeploy_app", shopProd("api"))
	token := planOf(t, text).PlanToken

	w.routes.mu.Lock()
	w.routes.source = gin.H{"activeMethod": "image", "imageSource": gin.H{"image": "evil/api:latest"}}
	w.routes.mu.Unlock()

	text, isErr := callTool(t, session, "apply_plan", map[string]any{"planToken": token})
	assert.True(t, isErr)
	assert.Contains(t, text, "changed since the plan was made")
	assert.Empty(t, w.routes.writes())
}

func TestRedeployRefusesWhatDoesNotFitTheSource(t *testing.T) {
	w := newWriteWorld(t)
	args := shopProd("api")
	args["repoRef"] = "main"
	text, isErr := callTool(t, w.session(t, "key1"), "plan_redeploy_app", args)
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	assert.Empty(t, plan.PlanToken)
	assert.Equal(t, nextNothing, plan.Next)
	assert.Contains(t, string(plan.Plan), "for an app built from a repository")
}

func installArgs() map[string]any {
	return map[string]any{"project": "shop", "env": "prod", "template": "wikijs", "name": "wiki",
		"params": map[string]any{"adminPassword": "hunter2-s3cr3t"}}
}

func TestInstallPlansEveryAppAndCreatesThemOnApply(t *testing.T) {
	w := newWriteWorld(t)
	session := w.session(t, "key1")

	text, isErr := callTool(t, session, "plan_install_app", installArgs())
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	var shown installPlan
	assert.NoError(t, json.Unmarshal(plan.Plan, &shown))
	if assert.Len(t, shown.Apps, 2) {
		assert.Equal(t, plannedApp{Name: "wiki-db", Key: "wiki-db", Kind: "dependency", Role: "db",
			Template: "postgres", Image: "postgres:17"}, shown.Apps[0])
	}
	assert.Empty(t, w.routes.writes(), "a plan creates nothing")

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"dependency:db":"n0"`)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/from-template", sent[0].Path)
		assert.Contains(t, sent[0].Body, "hunter2-s3cr3t", "the endpoint is sent the value")
	}
	for _, entry := range w.audit.entries {
		assert.NotContains(t, entry.Detail, "hunter2", "the audit log keeps parameter names only")
	}
}

func TestAnInstallThatWouldBeRefusedHasNoPlan(t *testing.T) {
	w := newWriteWorld(t)
	w.routes.preflight = gin.H{"issues": []gin.H{{"code": "ERR_DOMAIN_TAKEN", "detail": "wiki.test is served"}}}
	text, isErr := callTool(t, w.session(t, "key1"), "plan_install_app", installArgs())
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	assert.Empty(t, plan.PlanToken)
	assert.Contains(t, string(plan.Plan), "wiki.test is served")
}

func TestAnInstallOverOldDataSaysSo(t *testing.T) {
	w := newWriteWorld(t)
	w.routes.preflight["storage"] = []gin.H{{"app": "wiki-db", "appKey": "wiki-db", "isDatabase": true,
		"volume": gin.H{"name": "data"}, "path": "shop/prod/wiki-db"}}
	text, _ := callTool(t, w.session(t, "key1"), "plan_install_app", installArgs())
	plan := planOf(t, text)
	assert.NotEmpty(t, plan.PlanToken, "keeping the data is the person's choice to make")
	assert.Contains(t, string(plan.Plan), "password it was created with")
}

const apiPath = "projects/shop/envs/prod/apps/api"

func configArgs(doc string) map[string]any {
	args := shopProd("api")
	args["yaml"] = doc
	return args
}

func updatePlan(changes ...string) gin.H {
	return gin.H{"planHash": "h1", "nodes": []gin.H{
		{"path": "projects/shop/envs/prod", "kind": "env", "selected": false, "action": "unchanged"},
		{"path": apiPath, "kind": "app", "key": "api", "selected": true, "action": "update",
			"changes": changes, "restart": true},
	}}
}

func TestAConfigChangeIsTheImportsPlan(t *testing.T) {
	w := newWriteWorld(t)
	w.routes.importPlan = updatePlan("settings/envVars")
	session := w.session(t, "key1")
	doc := "name: API\nsettings:\n  envVars:\n    - {k: LOG_LEVEL, v: debug}\n"

	text, isErr := callTool(t, session, "plan_update_app_config", configArgs(doc))
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	var shown configPlan
	assert.NoError(t, json.Unmarshal(plan.Plan, &shown))
	assert.Equal(t, []string{"settings/envVars"}, shown.Changes)
	assert.True(t, shown.Restart)
	assert.Equal(t, apiPath+"\n"+doc, w.routes.validated,
		"the import is asked about the app alone, with the model's document in place of its own")
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"outcome":"updated"`)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/spec/import/apply", sent[0].Path)
		var body map[string]any
		assert.NoError(t, json.Unmarshal([]byte(sent[0].Body), &body))
		assert.Equal(t, "h1", body["planHash"])
		assert.NotEmpty(t, body["bundle"])
	}
}

func TestAConfigChangeRefusedAsMovedSaysPlanAgain(t *testing.T) {
	w := newWriteWorld(t)
	w.routes.importPlan = updatePlan("settings/envVars")
	session := w.session(t, "key1")
	text, _ := callTool(t, session, "plan_update_app_config", configArgs("name: API\n"))
	w.routes.planChanged = true
	text, isErr := callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.True(t, isErr)
	assert.Contains(t, text, "the app's configuration changed since the plan was made")
}

func TestAConfigChangeThatChangesNothingOrIsBlockedHasNoPlan(t *testing.T) {
	w := newWriteWorld(t)
	session := w.session(t, "key1")

	w.routes.importPlan = gin.H{"planHash": "h1", "nodes": []gin.H{
		{"path": apiPath, "kind": "app", "selected": true, "action": "unchanged"}}}
	text, _ := callTool(t, session, "plan_update_app_config", configArgs("name: API\n"))
	assert.Empty(t, planOf(t, text).PlanToken, "nothing to change")

	w.routes.importPlan = gin.H{"planHash": "h1", "nodes": []gin.H{
		{"path": apiPath, "kind": "app", "selected": true, "action": "update", "changes": []string{"x"},
			"issues": []gin.H{{"severity": "blocked", "code": "MOUNT_TARGET_DUPLICATED"}}}}}
	text, _ = callTool(t, session, "plan_update_app_config", configArgs("name: API\n"))
	plan := planOf(t, text)
	assert.Empty(t, plan.PlanToken, "blocked")
	assert.Contains(t, string(plan.Plan), "MOUNT_TARGET_DUPLICATED")
}

func TestAConfigDocumentIsOneMapping(t *testing.T) {
	w := newWriteWorld(t)
	for _, doc := range []string{"- a\n- b\n", "name: [unclosed\n", ""} {
		text, isErr := callTool(t, w.session(t, "key1"), "plan_update_app_config", configArgs(doc))
		assert.True(t, isErr, doc)
		assert.Contains(t, text, "yaml is", doc)
	}
}

func TestASchedJobIsPlannedWithItsRuns(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()
	w := newWriteWorld(t)
	session := w.session(t, "key1")
	args := shopProd("api")
	args["name"], args["cronExpr"], args["timeZone"] = "nightly cleanup", "0 2 * * *", "Asia/Ho_Chi_Minh"
	args["command"], args["timeout"] = "php artisan cleanup", "30m"

	text, isErr := callTool(t, session, "plan_create_sched_job", args)
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	var shown schedJobPlan
	assert.NoError(t, json.Unmarshal(plan.Plan, &shown))
	assert.Equal(t, []string{"2026-09-27T02:00:00+07:00 Sun", "2026-09-28T02:00:00+07:00 Mon"}, shown.NextRuns)
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": plan.PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"id":"j9"`)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/a1/sched-jobs", sent[0].Path)
		assert.JSONEq(t, `{"name":"nightly cleanup","jobType":"container-command","app":{"id":"a1"},"maxRetry":0,
			"schedule":{"cronExpr":"0 2 * * *","initialTime":"2026-09-26T19:00:00+07:00"},
			"command":{"command":"php artisan cleanup"},"timeout":"30m",
			"notification":{"successUseDefault":true,"failureUseDefault":true}}`, sent[0].Body)
	}
}

func TestASchedJobNeedsWhatItRuns(t *testing.T) {
	w := newWriteWorld(t)
	for name, change := range map[string]map[string]any{
		"no command":      {"command": ""},
		"both schedules":  {"interval": "1h"},
		"a bad timeout":   {"timeout": "soon"},
		"too many tries":  {"maxRetry": 11},
		"an unknown zone": {"timeZone": "Mars/Base"},
	} {
		args := shopProd("api")
		args["name"], args["cronExpr"], args["command"] = "job", "@daily", "true"
		for k, v := range change {
			args[k] = v
		}
		_, isErr := callTool(t, w.session(t, "key1"), "plan_create_sched_job", args)
		assert.True(t, isErr, name)
	}
}

func TestPromptsPlanWhenChangesAreAllowed(t *testing.T) {
	w := newWriteWorld(t)
	res, err := w.session(t, "key1").GetPrompt(context.Background(), &mcpsdk.GetPromptParams{Name: "debug_app",
		Arguments: map[string]string{"project": "shop", "env": "prod", "app": "api"}})
	if assert.NoError(t, err) {
		assert.Contains(t, res.Messages[0].Content.(*mcpsdk.TextContent).Text, "apply it only once I agree")
	}
	res, err = w.session(t, "reader").GetPrompt(context.Background(), &mcpsdk.GetPromptParams{Name: "debug_app",
		Arguments: map[string]string{"project": "shop", "env": "prod", "app": "api"}})
	if assert.NoError(t, err) {
		assert.Contains(t, res.Messages[0].Content.(*mcpsdk.TextContent).Text, "these tools only read")
	}
}
