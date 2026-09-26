package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

type fakeSwitch struct{ on bool }

func (f *fakeSwitch) IsEnabled(context.Context) (bool, error) { return f.on, nil }

// fakeKeys takes the key "key1" and nothing else.
type fakeKeys struct{}

func (fakeKeys) GetAPIKeyAuth(ctx *gin.Context) (*basedto.Auth, error) {
	if ctx.GetHeader("HIVEPAAS-API-KEY-ID") != "key1" {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}
	return testAuth(), nil
}

func (fakeKeys) RequestCtx(ctx *gin.Context) context.Context { return ctx.Request.Context() }

type fakeAudit struct {
	auditservice.Service
	mu      sync.Mutex
	entries []*auditservice.Entry
}

func (f *fakeAudit) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, entry)
	return nil
}

// mcpWorld is the router with a fake project endpoint and the MCP endpoint on it.
type mcpWorld struct {
	url    string
	sw     *fakeSwitch
	audit  *fakeAudit
	denied bool
}

func newMCPWorld(t *testing.T) *mcpWorld {
	t.Helper()
	w := &mcpWorld{sw: &fakeSwitch{on: true}, audit: &fakeAudit{}}
	engine := testEngine(func(api *gin.RouterGroup, auth *authhandler.Handler) {
		api.GET("/projects", func(ctx *gin.Context) {
			if _, err := auth.GetCurrentAuth(ctx, authhandler.NoAccessCheck); err != nil || w.denied {
				ctx.JSON(http.StatusForbidden, gin.H{"title": "Forbidden", "status": 403})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"data": []gin.H{
				{"id": "p1", "key": "shop", "name": "Shop", "envs": []gin.H{{"id": "e1", "name": "dev"}}},
			}})
		})
	})
	services := &Services{Auth: fakeKeys{}, Switch: w.sw, Audit: w.audit}
	endpoint := NewEndpoint(services, NewDispatcher(engine, "/api"), Tools())
	engine.Any("/api/mcp", endpoint.Serve)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	w.url = server.URL + "/api/mcp"
	return w
}

type headerTransport struct{ key string }

func (h headerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if h.key != "" {
		r.Header.Set("HIVEPAAS-API-KEY-ID", h.key)
	}
	return http.DefaultTransport.RoundTrip(r)
}

func (w *mcpWorld) connect(t *testing.T, key string) (*mcpsdk.ClientSession, error) {
	t.Helper()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint: w.url, HTTPClient: &http.Client{Transport: headerTransport{key: key}},
	}, nil)
	if err == nil {
		t.Cleanup(func() { _ = session.Close() })
	}
	return session, err
}

func TestEndpointIsOffUntilEnabled(t *testing.T) {
	w := newMCPWorld(t)
	w.sw.on = false
	resp, err := http.Post(w.url, "application/json", strings.NewReader("{}"))
	if assert.NoError(t, err) {
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
	_, err = w.connect(t, "key1")
	assert.Error(t, err)
}

func TestEndpointTakesAnAPIKey(t *testing.T) {
	w := newMCPWorld(t)
	resp, err := http.Post(w.url, "application/json", strings.NewReader("{}"))
	if assert.NoError(t, err) {
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
	_, err = w.connect(t, "wrong")
	assert.Error(t, err)

	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	tools, err := session.ListTools(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, tools.Tools, len(Tools()))
}

// A call reads through the router as the caller, and is audited.
func TestAToolCallIsAnsweredAndAudited(t *testing.T) {
	w := newMCPWorld(t)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "list_projects", Arguments: map[string]any{},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.False(t, res.IsError)
	assert.Contains(t, res.Content[0].(*mcpsdk.TextContent).Text, `"key":"shop"`)
	assert.Contains(t, res.Content[0].(*mcpsdk.TextContent).Text, `"envs":["dev"]`)

	if assert.Len(t, w.audit.entries, 1) {
		entry := w.audit.entries[0]
		assert.Equal(t, base.AuditLogTypeMCPToolCall, entry.Type)
		assert.Equal(t, base.AuditLogSourceMCP, entry.Source)
		assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
		assert.Equal(t, "list_projects", entry.ResName)
		assert.Equal(t, "u1", entry.Auth.UserID())
	}
}

// What the endpoint refuses, the tool reports as a tool error the model reads,
// and the refusal is audited.
func TestARefusalIsAToolErrorAndAudited(t *testing.T) {
	w := newMCPWorld(t)
	w.denied = true
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "list_projects", Arguments: map[string]any{},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*mcpsdk.TextContent).Text, "not permitted")
	if assert.Len(t, w.audit.entries, 2) {
		assert.Equal(t, base.AuditLogResultDenied, w.audit.entries[1].Result)
	}
}

func TestEveryToolIsReadAndDescribed(t *testing.T) {
	names := map[string]bool{}
	for _, tool := range Tools() {
		assert.Equal(t, KindRead, tool.Kind, tool.Name)
		assert.NotEmpty(t, tool.Title, tool.Name)
		assert.True(t, strings.HasSuffix(tool.Description, "."), "%s: describe it in a sentence", tool.Name)
		assert.False(t, names[tool.Name], "%s is registered twice", tool.Name)
		names[tool.Name] = true
	}
}
