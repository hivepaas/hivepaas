package mcp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// appRoutes is one env, shop/prod, with two apps the caller sees - api, which
// owns api-db, and worker - and one it does not. Wherever an answer may hold a
// secret the tools must not pass on, it holds s3cr3t-<n>.
type appRoutes struct {
	logs       []logFrame
	logsTail   string
	exportMode string
}

func (r *appRoutes) add(api *gin.RouterGroup) {
	env := api.Group("/projects/p1/prod")
	env.GET("/apps", func(ctx *gin.Context) {
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
	env.POST("/apps/a1/spec/export", func(ctx *gin.Context) {
		var body struct {
			SecretsMode string `json:"secretsMode"`
		}
		_ = ctx.ShouldBindJSON(&body)
		r.exportMode = body.SecretsMode
		password := ""
		if body.SecretsMode != "omit" {
			password = "s3cr3t-3"
		}
		ctx.Data(http.StatusOK, "application/gzip", testBundle(map[string]string{
			"spec.yaml": "kind: bundle\n",
			"projects/shop/envs/prod.yaml": "kind: env\nproject: shop\nenv: prod\napps:\n" +
				"  api:\n    name: API\n    settings:\n      envVars:\n        - {k: DB_PASSWORD, v: '" + password + "'}\n" +
				"  worker:\n    name: Worker\n",
		}))
	})
}

func testBundle(files map[string]string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	archive := tar.NewWriter(gz)
	for name, content := range files {
		_ = archive.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(content))})
		_, _ = archive.Write([]byte(content))
	}
	_ = archive.Close()
	_ = gz.Close()
	return buf.Bytes()
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

func TestListApps(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("list_apps", map[string]any{"project": "shop", "env": "prod"})
	assert.False(t, isErr, text)
	var out appList
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, "shop", out.Project)
	if assert.Len(t, out.Apps, 3) {
		assert.Equal(t, "api", out.Apps[0].Key)
		assert.Equal(t, 1, *out.Apps[0].Running)
		assert.Equal(t, 2, *out.Apps[0].Desired)
		assert.Equal(t, "api-db", out.Apps[1].Key)
		assert.Equal(t, "api", out.Apps[1].Owner)
		assert.Equal(t, "worker", out.Apps[2].Key)
		assert.Nil(t, out.Apps[2].Running)
	}
}

func TestGetApp(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("get_app", shopProd("api"))
	assert.False(t, isErr, text)
	var out appDetail
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []string{"api-db"}, out.Children)
	assert.Equal(t, []string{"https://api.shop.test"}, out.Links)
	assert.Equal(t, &appSource{Method: "image", Image: "shop/api:2"}, out.Source)
	assert.Equal(t, "https://git.test/shop/api.git", withoutUserinfo("https://ada:s3cr3t-4@git.test/shop/api.git"))
	if assert.Len(t, out.Deployments, 2) {
		assert.Equal(t, "failed", out.Deployments[0].Status)
		assert.Equal(t, "pull access denied for shop/api", out.Deployments[0].Error)
		assert.Equal(t, "user", out.Deployments[0].Trigger)
	}
}

func TestGetAppStatusIsNewestFirst(t *testing.T) {
	_, call := appSession(t)
	text, isErr := call("get_app_status", shopProd("api"))
	assert.False(t, isErr, text)
	var out appStatus
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	if assert.Len(t, out.Tasks, 2) {
		assert.Equal(t, taskItem{ID: "fedcba987654", Slot: 1, Node: "n2", State: "rejected", Desired: "running",
			Error: "No such image: shop/api:2", At: time.Date(2026, 9, 26, 8, 0, 0, 0, time.UTC)}, out.Tasks[0])
		assert.Equal(t, "failed", out.Tasks[1].State)
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
	var out appLogs
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, []string{"2026-09-26T08:00:20.000Z [stderr] Error: connection refused 20"}, out.Lines)
	assert.Equal(t, 2, *out.Matched)
	assert.Equal(t, 300, out.Scanned)

	args["grep"] = `/refused (1|2)0$/`
	text, _ = call("get_app_logs", args)
	assert.Contains(t, text, "connection refused 20")

	args = shopProd("api")
	args["tail"] = 3
	text, _ = call("get_app_logs", args)
	assert.Equal(t, "3", routes.logsTail)
	out = appLogs{}
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Len(t, out.Lines, 3)
	assert.Nil(t, out.Matched)

	args["tail"] = 501
	text, isErr = call("get_app_logs", args)
	assert.True(t, isErr)
	assert.Contains(t, text, "tail")
}

func TestGetAppConfigIsTheAppsDocument(t *testing.T) {
	routes, call := appSession(t)
	text, isErr := call("get_app_config", shopProd("api"))
	assert.False(t, isErr, text)
	assert.Equal(t, "omit", routes.exportMode)
	var out appConfig
	assert.NoError(t, json.Unmarshal([]byte(text), &out))
	assert.Equal(t, "name: API\nsettings:\n  envVars:\n    - {k: DB_PASSWORD, v: ''}\n", out.YAML)
}

// Whatever an app holds, no tool that reads it answers a secret value: they
// read only what they show, and export in the mode that omits secrets.
func TestNoToolOutputCarriesASecretValue(t *testing.T) {
	_, call := appSession(t)
	for _, tool := range []string{"list_apps", "get_app", "get_app_status", "get_app_config"} {
		args := shopProd("api")
		if tool == "list_apps" {
			delete(args, "app")
		}
		text, isErr := call(tool, args)
		assert.False(t, isErr, "%s: %s", tool, text)
		assert.NotContains(t, text, "s3cr3t", tool)
	}
}

func TestParseSince(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	for in, want := range map[string]time.Time{
		"":                     {},
		"2h":                   now.Add(-2 * time.Hour),
		"2026-09-26T08:00:00Z": time.Date(2026, 9, 26, 8, 0, 0, 0, time.UTC),
	} {
		got, err := parseSince(in, now)
		assert.NoError(t, err, in)
		assert.True(t, want.Equal(got), in)
	}
	for _, in := range []string{"yesterday", "-1h"} {
		_, err := parseSince(in, now)
		assert.ErrorAs(t, err, new(*InputError), in)
	}
}
