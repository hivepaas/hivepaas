package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// appRoutes is one env, shop/prod, with two apps the caller sees - api, which
// owns api-db, and worker - and one it does not. Wherever an answer may hold a
// secret the tools must not pass on, it holds s3cr3t-<n>.
type appRoutes struct {
	logs     []logFrame
	logsTail string
	// query is the query the last list endpoint was asked with.
	query url.Values
}

func (r *appRoutes) add(api *gin.RouterGroup) {
	env := api.Group("/projects/p1/prod")
	env.GET("/apps", func(ctx *gin.Context) {
		r.query = ctx.Request.URL.Query()
		// As the permission layer does: the app the caller may not see is not listed.
		ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{
			{"id": "a1", "key": "api", "name": "API", "status": "active", "engine": "",
				"stats":            gin.H{"runningTasks": 1, "desiredTasks": 2},
				"logicalChildApps": []gin.H{{"id": "a3", "key": "api-db", "name": "API DB", "status": "active"}}},
			{"id": "a2", "key": "worker", "name": "Worker", "status": "active"},
		}})
	})
	env.GET("/apps/a1", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"id": "a1", "key": "api", "name": "API", "status": "active", "note": "the shop's API",
			"accessLinks":      []string{"https://api.shop.test"},
			"stats":            gin.H{"runningTasks": 1, "desiredTasks": 2},
			"logicalChildApps": []gin.H{{"id": "a3", "key": "api-db"}},
			// Not what get_app reads; here to show it is not passed on.
			"settings": gin.H{"envVars": "DB_PASSWORD=s3cr3t-1"},
		}})
	})
	env.GET("/apps/a1/deployments", func(ctx *gin.Context) {
		r.query = ctx.Request.URL.Query()
		ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{{
			"id": "d2", "status": "failed", "createdAt": "2026-09-26T08:00:00Z",
			"trigger": gin.H{"source": "user"},
			"output":  gin.H{"error": "pull access denied for shop/api"},
			"settings": gin.H{"activeMethod": "image", "imageSource": gin.H{"image": "shop/api:2"},
				"command": "serve --token s3cr3t-2"},
		}, {
			"id": "d1", "status": "done", "createdAt": "2026-09-25T08:00:00Z",
			"settings": gin.H{"activeMethod": "image", "imageSource": gin.H{"image": "shop/api:1"},
				"repoSource": gin.H{"repoURL": "https://ada:s3cr3t-4@git.test/shop/api.git"}},
		}}})
	})
	env.GET("/apps/a1/service-tasks", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{
			{"id": "0123456789abcdef", "slot": 1, "node": gin.H{"hostname": "n1"}, "desiredState": "shutdown",
				"status": gin.H{"timestamp": "2026-09-26T07:00:00Z", "state": "failed", "err": "task: non-zero exit (1)"}},
			{"id": "fedcba9876543210", "slot": 1, "node": gin.H{"hostname": "n2"}, "desiredState": "running",
				"status": gin.H{"timestamp": "2026-09-26T08:00:00Z", "state": "rejected",
					"err": "No such image: shop/api:2"}},
		}})
	})
	env.GET("/apps/a1/logs", func(ctx *gin.Context) {
		r.logsTail = ctx.Query("tail")
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"logs": r.logs}})
	})
	env.GET("/apps/a1/kind-settings", func(ctx *gin.Context) {
		// As the endpoint answers without revealSecrets: the password masked.
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"category": "database", "engine": "postgres",
			"database":     gin.H{"dbName": "shop", "username": "shop", "password": "********"},
			"secretMasked": true, "updateVer": 1}})
	})
}

func appSession(t *testing.T) (*appRoutes, func(name string, args map[string]any) (string, bool)) {
	t.Helper()
	routes := &appRoutes{}
	w := newMCPWorld(t, routes.add)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return routes, func(name string, args map[string]any) (string, bool) {
		return callTool(t, session, name, args)
	}
}

func shopProd(app string) map[string]any {
	return map[string]any{"project": "shop", "env": "prod", "app": app}
}

func TestResolveTakesKeysAndNames(t *testing.T) {
	_, call := appSession(t)
	for _, args := range []map[string]any{
		shopProd("api"),
		shopProd("API"),
		shopProd("a1"),
		{"project": "p1", "env": "prod", "app": "api"},
	} {
		text, isErr := call("get_app", args)
		assert.False(t, isErr, "%v: %s", args, text)
		assert.Contains(t, text, `"key":"api"`, args)
	}
	// An app another owns is found too.
	text, isErr := call("get_app_status", shopProd("api-db"))
	assert.True(t, isErr) // the fake has no tasks for it
	assert.Contains(t, text, "not found", text)
}

func TestResolveRefusesAmbiguity(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("list_apps", map[string]any{"project": "Shop", "env": "prod"})
	assert.True(t, isErr)
	assert.Contains(t, text, `"Shop" names 2 of the projects`)
	assert.Contains(t, text, "shop (Shop)")
	assert.Contains(t, text, "shop-eu (Shop)")
}

func TestResolveHidesWhatTheCallerCannotSee(t *testing.T) {
	_, call := appSession(t)
	hidden, isErr := call("get_app", shopProd("billing"))
	assert.True(t, isErr)
	missing, _ := call("get_app", shopProd("nothing"))
	assert.Equal(t, strings.ReplaceAll(missing, "nothing", "billing"), hidden)

	text, isErr := call("list_apps", map[string]any{"project": "shop", "env": "staging"})
	assert.True(t, isErr)
	assert.Contains(t, text, `no env of shop "staging"`)
}

// A read tool answers what its endpoint answers, decoded into the endpoint's
// own response type: {meta, data}.
func TestListApps(t *testing.T) {
	routes, call := appSession(t)
	text, isErr := call("list_apps", map[string]any{"project": "shop", "env": "prod", "getStats": true})
	assert.False(t, isErr, text)
	assert.Equal(t, url.Values{"getStats": {"true"}}, routes.query)
	var out appdto.ListAppResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data, 2) {
		assert.Equal(t, "api", out.Data[0].Key)
		assert.Equal(t, 1, out.Data[0].Stats.RunningTasks)
		assert.Equal(t, 2, out.Data[0].Stats.DesiredTasks)
		assert.Equal(t, "api-db", out.Data[0].LogicalChildApps[0].Key)
		assert.Equal(t, "worker", out.Data[1].Key)
	}
}

func TestGetApp(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("get_app", shopProd("api"))
	assert.False(t, isErr, text)
	var out appdto.GetAppResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, "the shop's API", out.Data.Note)
	assert.Equal(t, "api-db", out.Data.LogicalChildApps[0].Key)
	assert.Equal(t, "https://git.test/shop/api.git", withoutUserinfo("https://ada:s3cr3t-4@git.test/shop/api.git"))
}

func TestGetAppStatus(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("get_app_status", shopProd("api"))
	assert.False(t, isErr, text)
	var out appsettingsdto.GetAppServiceTasksResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data, 2) {
		assert.Equal(t, "No such image: shop/api:2", out.Data[1].Status.Err)
	}
}

func TestListAppDeployments(t *testing.T) {
	routes, call := appSession(t)
	args := shopProd("api")
	args["status"] = []any{"failed", "done"}
	args["pageLimit"] = 5
	text, isErr := call("list_app_deployments", args)
	assert.False(t, isErr, text)
	assert.Equal(t, url.Values{"status": {"failed,done"}, "pageLimit": {"5"}}, routes.query,
		"a list goes as the handler reads one: its values joined by commas")
	var out appdeploymentdto.ListDeploymentResp
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Data, 2) {
		assert.Equal(t, "pull access denied for shop/api", out.Data[0].Output.Error)
	}
}

func TestLogsGrepBeforeTail(t *testing.T) {
	routes, call := appSession(t)
	base := time.Date(2026, 9, 26, 8, 0, 0, 0, time.UTC)
	for i := range 300 {
		frame := logFrame{Type: "out", Data: fmt.Sprintf("GET /health %d\n", i), Ts: base.Add(time.Duration(i) * time.Second)}
		if i == 10 || i == 20 {
			frame = logFrame{Type: "err", Data: fmt.Sprintf("Error: connection refused %d", i), Ts: frame.Ts}
		}
		routes.logs = append(routes.logs, frame)
	}

	args := shopProd("api")
	args["tail"], args["grep"] = 1, "ERROR"
	text, isErr := call("get_app_logs", args)
	assert.False(t, isErr, text)
	assert.Equal(t, "5000", routes.logsTail, "a grep reads all it can, then tails")
	var out logsAnswer
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []string{"2026-09-26T08:00:20.000Z [stderr] Error: connection refused 20"}, out.Lines)
	assert.Equal(t, 2, *out.Matched)
	assert.Equal(t, 300, out.Answered)

	args["grep"] = `/refused (1|2)0$/`
	text, _ = call("get_app_logs", args)
	assert.Contains(t, text, "connection refused 20")

	args = shopProd("api")
	args["tail"] = 3
	text, _ = call("get_app_logs", args)
	assert.Equal(t, "3", routes.logsTail)
	out = logsAnswer{}
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Len(t, out.Lines, 3)
	assert.Nil(t, out.Matched)

	args["tail"] = 5001
	text, isErr = call("get_app_logs", args)
	assert.True(t, isErr)
	assert.Contains(t, text, "tail")
}

// Whatever an app holds, no tool that reads it answers a secret value: they
// read only what they show, and export in the mode that omits secrets.
func TestNoToolOutputCarriesASecretValue(t *testing.T) {
	_, call := appSession(t)
	for _, tool := range []string{"list_apps", "get_app", "get_app_settings"} {
		args := shopProd("api")
		switch tool {
		case "list_apps":
			delete(args, "app")
		case "get_app_settings":
			args["kind"] = "kind"
		}
		text, isErr := call(tool, args)
		assert.False(t, isErr, "%s: %s", tool, text)
		assert.NotContains(t, text, "s3cr3t", tool)
	}
}

// The log endpoint answers several containers' lines each in order but not
// with one another: the tail keeps the newest, whichever container they are of.
func TestLogsAreTailedInTimeOrder(t *testing.T) {
	at := func(day int) time.Time { return time.Date(2026, 9, day, 0, 0, 0, 0, time.UTC) }
	frames := []logFrame{
		{Type: "out", Data: "second container, newest", Ts: at(26)},
		{Type: "out", Data: "first container, oldest", Ts: at(20)},
		{Type: "out", Data: "first container, later", Ts: at(24)},
	}
	out := makeLogsAnswer(frames, 2, nil)
	assert.Equal(t, []string{
		"2026-09-24T00:00:00.000Z first container, later",
		"2026-09-26T00:00:00.000Z second container, newest",
	}, out.Lines)
}
