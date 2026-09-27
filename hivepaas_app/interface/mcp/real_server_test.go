package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// TestAgainstARealServer calls every tool once, through the SDK's own client,
// against a running HivePaaS with the MCP server enabled: the real router, the
// real handlers and their permission checks, the real database. It runs only
// when HP_TEST_MCP_URL is set, with
//
//	HP_TEST_MCP_URL     the endpoint, such as http://localhost:8080/_/mcp
//	HP_TEST_MCP_KEY     an API key as <keyId>:<secret>
//	HP_TEST_MCP_APP     an app to read, as <project>/<env>/<app>
//	HP_TEST_MCP_WRITE   1 to also plan and apply a restart of that app: the
//	                    server must allow changes and the key must execute
func TestAgainstARealServer(t *testing.T) {
	endpoint := os.Getenv("HP_TEST_MCP_URL")
	if endpoint == "" {
		t.Skip("set HP_TEST_MCP_URL, HP_TEST_MCP_KEY and HP_TEST_MCP_APP to run against a running server")
	}
	key := os.Getenv("HP_TEST_MCP_KEY")
	app := strings.Split(os.Getenv("HP_TEST_MCP_APP"), "/")
	if !assert.Len(t, app, 3, "HP_TEST_MCP_APP is <project>/<env>/<app>") {
		t.FailNow()
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "hivepaas-test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: &http.Client{Transport: bearerTransport{token: key}},
	}, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() { _ = session.Close() }()

	listed, err := session.ListTools(context.Background(), nil)
	if assert.NoError(t, err) {
		assert.GreaterOrEqual(t, len(listed.Tools), readToolCountOf(Tools()))
	}

	inApp := map[string]any{"project": app[0], "env": app[1], "app": app[2]}
	inEnv := map[string]any{"project": app[0], "env": app[1]}
	calls := []struct {
		tool string
		args map[string]any
		// mayRefuse is a call the endpoint may refuse and still have been
		// reached: a key limited to reading, a template parameter the test
		// cannot know, a feature this installation has off.
		mayRefuse bool
	}{
		{"list_projects", nil, false},
		{"list_apps", merge(inEnv, map[string]any{"getStats": true, "getChildApps": true}), false},
		{"get_app", merge(inApp, map[string]any{"getStats": true}), false},
		{"get_app_status", inApp, false},
		{"list_app_deployments", merge(inApp, map[string]any{"pageLimit": 5}), false},
		{"get_app_logs", merge(inApp, map[string]any{"tail": 20, "grep": "/err|warn/"}), false},
		{"search_app_logs", merge(inApp, map[string]any{"limit": 20}), true}, // refused when logging is off
		{"get_app_settings", merge(inApp, map[string]any{"kind": "env-vars"}), false},
		{"list_attention", nil, false},
		{"list_tasks", merge(inEnv, map[string]any{"pageLimit": 5, "sort": "-createdAt"}), false},
		{"list_nodes", nil, false},
		{"get_template_catalog", nil, false},
		{"search_templates", map[string]any{"search": "postgres", "pageLimit": 5}, false},
		{"get_template", map[string]any{"template": "postgres"}, false},
		{"list_template_image_tags", map[string]any{"template": "postgres"}, true}, // reads the registry
		{"preflight_install", merge(inEnv, map[string]any{"template": "redis", "name": "mcp-preflight"}), true},
		{"list_sched_jobs", nil, false},
		{"explain_schedule", map[string]any{"cronExpr": "0 2 * * *", "count": 3,
			"initialTime": "2026-09-27T00:00:00+07:00"}, false},
	}
	for _, c := range calls {
		res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: c.tool, Arguments: c.args})
		if !assert.NoError(t, err, c.tool) {
			continue
		}
		text := res.Content[0].(*mcpsdk.TextContent).Text
		if res.IsError && c.mayRefuse {
			t.Logf("%s: refused by the endpoint: %s", c.tool, text)
			continue
		}
		assert.False(t, res.IsError, "%s: %s", c.tool, text)
		assert.LessOrEqual(t, len(text), MaxToolOutput, c.tool)
		t.Logf("%s: %d bytes", c.tool, len(text))
	}

	if os.Getenv("HP_TEST_MCP_WRITE") == "1" {
		restartForReal(t, session, inApp)
	}
}

// restartForReal plans a restart of the app and applies it, as an assistant would once agreed.
func restartForReal(t *testing.T, session *mcpsdk.ClientSession, inApp map[string]any) {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: "plan_restart_app",
		Arguments: inApp})
	if !assert.NoError(t, err) || !assert.False(t, res.IsError, "plan_restart_app") {
		return
	}
	var plan planResult[json.RawMessage]
	if !assert.NoError(t, json.Unmarshal([]byte(res.Content[0].(*mcpsdk.TextContent).Text), &plan)) ||
		!assert.NotEmpty(t, plan.PlanToken) {
		return
	}
	t.Logf("plan: %s", plan.Plan)
	res, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: "apply_plan",
		Arguments: map[string]any{"planToken": plan.PlanToken}})
	if assert.NoError(t, err) {
		text := res.Content[0].(*mcpsdk.TextContent).Text
		assert.False(t, res.IsError, text)
		t.Logf("applied: %s", text)
	}
}

func readToolCountOf(tools []Tool) int {
	n := 0
	for _, tool := range tools {
		if tool.Kind == KindRead {
			n++
		}
	}
	return n
}

type bearerTransport struct{ token string }

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(r)
}

func merge(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
