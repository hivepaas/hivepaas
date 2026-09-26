package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

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
	routes := &writeRoutes{source: gin.H{"activeMethod": "image", "imageSource": gin.H{"image": "shop/api:1.25"}}}
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
		case entry.ResName == "apply_plan" && strings.Contains(entry.Detail, "applies"):
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
