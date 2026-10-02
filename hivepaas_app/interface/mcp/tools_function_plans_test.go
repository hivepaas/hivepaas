package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// savedFunction is api's (a1) deployment settings: a function with inline code.
var savedFunction = gin.H{"activeMethod": "function", "functionSource": gin.H{
	"runtime": "node24", "entrypoint": gin.H{"file": "index.js", "handler": "default"},
	"code": gin.H{"inline": gin.H{"files": []gin.H{
		{"path": "index.js", "content": "export default () => ({ body: 'saved' })"},
	}}},
}}

// functionRoutes answers what the function tools read and send: api (a1) is a
// function with inline code - its settings answered by writeRoutes - and worker
// (a2) an image app.
func functionRoutes(api *gin.RouterGroup) {
	env := api.Group("/projects/p1/prod")
	env.POST("/apps/function", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": "f9", "deploymentId": "d9", "taskId": "t9"}})
	})
	env.GET("/apps/a2/deployment-settings", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"activeMethod": "image",
			"imageSource": gin.H{"image": "nginx"}}})
	})
	env.POST("/apps/a1/function/test-run", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"outcome": "ok", "status": 200, "body": []byte(`{"hello":"Ada"}`), "durationMs": 1.5,
			"logs": `{"hp":"invocation","outcome":"ok"}`, "exitCode": 0, "librariesBuilt": true,
			"lockFiles": []gin.H{{"path": "package-lock.json", "content": "{}"}},
		}})
	})
	env.GET("/apps/a1/function-metrics", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"available": true, "range": ctx.Query("range"),
			"totals": gin.H{"calls": 42, "failed": 1}}})
	})
}

func newFunctionWorld(t *testing.T) *writeWorld {
	t.Helper()
	routes := &writeRoutes{source: savedFunction}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, functionRoutes)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

func sentBody(t *testing.T, sent sentRequest) map[string]any {
	t.Helper()
	var body map[string]any
	assert.NoError(t, json.Unmarshal([]byte(sent.Body), &body))
	return body
}

// A function is created from its files, deployed at once, at a domain given;
// nothing is sent until the plan is applied.
func TestCreatingAFunction(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_create_function", map[string]any{
		"project": "shop", "env": "prod", "name": "Resize", "runtime": "node24",
		"entrypoint": "index.ts", "domain": " Resize.Shop.Test. ",
		"files": []any{
			map[string]any{"path": "./index.ts", "content": "export default (): object => ({ ok: true })"},
			map[string]any{"path": "package.json", "content": `{"type":"module"}`},
		},
	})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "index.ts / default")
	assert.Contains(t, text, "https://resize.shop.test")
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "f9")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/function", sent[0].Path)
		body := sentBody(t, sent[0])
		assert.Equal(t, "Resize", body["name"])
		assert.Equal(t, "active", body["status"])
		assert.Equal(t, "resize.shop.test", body["domain"])
		source := body["source"].(map[string]any)
		assert.Equal(t, "node24", source["runtime"])
		assert.Equal(t, "index.ts", source["entrypoint"].(map[string]any)["file"])
		files := source["code"].(map[string]any)["inline"].(map[string]any)["files"].([]any)
		assert.Equal(t, "index.ts", files[0].(map[string]any)["path"], "the path as the API cleans it")
	}
}

// What the API would refuse is refused before a plan is made: a name taken, an
// entrypoint the runtime cannot load, no code.
func TestAFunctionThatWouldBeRefusedIsNotPlanned(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	file := []any{map[string]any{"path": "main.py", "content": "def handler(req, ctx): pass"}}
	for name, c := range map[string]struct {
		args map[string]any
		want string
	}{
		"a name taken": {map[string]any{"name": "worker", "runtime": "python313", "files": file},
			"has an app named"},
		"a JavaScript entrypoint for Python": {map[string]any{"name": "fn", "runtime": "python313",
			"entrypoint": "index.js", "files": file}, "ERR_VLD_FUNCTION_ENTRYPOINT_INVALID"},
		"a runtime that is none": {map[string]any{"name": "fn", "runtime": "ruby34", "files": file},
			"source.runtime"},
		"no code": {map[string]any{"name": "fn", "runtime": "python313"}, `"files"`},
	} {
		c.args["project"], c.args["env"] = "shop", "prod"
		text, isErr := callTool(t, session, "plan_create_function", c.args)
		assert.True(t, isErr, name)
		assert.Contains(t, text, c.want, name)
	}
	assert.Empty(t, w.routes.writes())
}

// A test run takes the function's saved code unless given files, sends the
// request as the endpoint reads it, and answers a text body as text.
func TestTestRunningAFunction(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_test_run_function", map[string]any{
		"project": "shop", "env": "prod", "app": "api", "path": "hello",
		"query": map[string]any{"name": "Ada"}, "headers": map[string]any{"X-Trace": "1"},
	})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "the function's saved code")
	assert.Contains(t, text, "GET /hello")
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `{\"hello\":\"Ada\"}`, "the body, as text")
	assert.Contains(t, text, "package-lock.json")
	assert.NotContains(t, text, `"content"`, "a lock file is named, not sent back")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/a1/function/test-run", sent[0].Path)
		body := sentBody(t, sent[0])
		files := body["code"].(map[string]any)["files"].([]any)
		assert.Contains(t, files[0].(map[string]any)["content"], "saved")
		req := body["request"].(map[string]any)
		assert.Equal(t, "GET", req["method"])
		assert.Equal(t, "/hello", req["path"])
		assert.Equal(t, []any{"Ada"}, req["query"].(map[string]any)["name"])
		assert.Equal(t, []any{"1"}, req["headers"].(map[string]any)["x-trace"], "a header's name in lower case")
	}
}

func TestTestRunningCodeGiven(t *testing.T) {
	w := newFunctionWorld(t)
	text, isErr := callTool(t, w.session(t, "key1"), "plan_test_run_function", map[string]any{
		"project": "shop", "env": "prod", "app": "api", "method": "post", "body": `{"n":1}`,
		"files": []any{map[string]any{"path": "index.js", "content": "export default () => 'fixed'"}},
	})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "the files given")
	assert.Contains(t, text, "POST /")
}

// Only a function has code to run.
func TestTestRunningAnAppThatIsNoFunction(t *testing.T) {
	w := newFunctionWorld(t)
	text, isErr := callTool(t, w.session(t, "key1"), "plan_test_run_function",
		map[string]any{"project": "shop", "env": "prod", "app": "worker"})
	assert.True(t, isErr)
	assert.Contains(t, text, "worker is not a function")
}

func TestAFunctionsMetrics(t *testing.T) {
	w := newFunctionWorld(t)
	text, isErr := callTool(t, w.session(t, "reader"), "get_function_metrics",
		map[string]any{"project": "shop", "env": "prod", "app": "api", "range": "6h"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"calls":42`)
	assert.Contains(t, text, `"range":"6h"`)
}

func TestTheFunctionsGuideAndPrompt(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	res, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: functionsURI})
	if assert.NoError(t, err) && assert.Len(t, res.Contents, 1) {
		for _, s := range []string{"bun1", "plan_test_run_function", "import type", "requirements.txt"} {
			assert.Contains(t, res.Contents[0].Text, s)
		}
	}

	args := map[string]string{"project": "shop", "env": "prod", "what": "resizes images", "runtime": "go127"}
	text := promptText(t, session, "deploy_function", args)
	for _, s := range []string{"resizes images", "go127", functionsURI, "plan_create_function",
		"plan_test_run_function", "apply it only once I agree"} {
		assert.Contains(t, text, s)
	}
	text = promptText(t, w.session(t, "reader"), "deploy_function", args)
	assert.Contains(t, text, "Change nothing: this key may not")
}

// A scheduled job of a function sends it a request: jobType function-invoke,
// set by the tool, with no command.
func TestSchedulingAFunctionsCall(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_create_sched_job", map[string]any{
		"project": "shop", "env": "prod", "app": "api", "name": "nightly report",
		"schedule":       map[string]any{"cronExpr": "0 3 * * *"},
		"functionInvoke": map[string]any{"method": "post", "path": "report", "body": `{"day":"yesterday"}`},
	})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "function-invoke")
	assert.Empty(t, w.routes.writes())

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "/projects/p1/prod/apps/a1/sched-jobs", sent[0].Path)
		body := sentBody(t, sent[0])
		assert.Equal(t, "function-invoke", body["jobType"])
		assert.Nil(t, body["command"])
		invoke := body["functionInvoke"].(map[string]any)
		assert.Equal(t, "post", invoke["method"], "the endpoint puts it in upper case")
		assert.JSONEq(t, `{"day":"yesterday"}`, invoke["body"].(string))
	}
}

// A function's call is only for a function, runs no command, and a job is one
// or the other.
func TestAFunctionsCallThatCannotBeScheduled(t *testing.T) {
	w := newFunctionWorld(t)
	session := w.session(t, "key1")
	schedule := map[string]any{"cronExpr": "@daily"}
	call := map[string]any{"path": "/"}
	for name, c := range map[string]struct {
		args map[string]any
		want string
	}{
		"an app that is no function": {map[string]any{"app": "worker", "functionInvoke": call},
			"worker is not a function"},
		"a command too": {map[string]any{"app": "api", "functionInvoke": call,
			"command": map[string]any{"command": "echo hi"}}, "not both"},
		"neither": {map[string]any{"app": "api"}, "command is required"},
	} {
		c.args["project"], c.args["env"], c.args["name"], c.args["schedule"] = "shop", "prod", "job", schedule
		text, isErr := callTool(t, session, "plan_create_sched_job", c.args)
		assert.True(t, isErr, name)
		assert.Contains(t, text, c.want, name)
	}
	assert.Empty(t, w.routes.writes())
}
