package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reqinfo"
)

func TestRemoteAddrOf(t *testing.T) {
	tests := []struct{ in, want string }{
		{"10.0.0.9:5555", "10.0.0.9"},
		{"[::1]:5555", "::1"},
		{"10.0.0.9", "10.0.0.9"}, // no port, keep it as it is
		{"", ""},
	}
	for _, tt := range tests {
		if got := remoteAddrOf(tt.in); got != tt.want {
			t.Errorf("remoteAddrOf(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRequestCtxCarriesRequestInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &BaseHandler{}

	var first, second *reqinfo.RequestInfo
	engine := gin.New()
	// No proxy is trusted, matching the server's own default.
	if err := engine.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	engine.GET("/", func(c *gin.Context) {
		first = reqinfo.From(h.RequestCtx(c))
		second = reqinfo.From(h.RequestCtx(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:5555"
	req.Header.Set("User-Agent", "probe/1.0")
	// A caller volunteering someone else's address. With no proxy trusted it must
	// not become the address recorded against them.
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	engine.ServeHTTP(httptest.NewRecorder(), req)

	if first == nil {
		t.Fatal("no request info reached the context")
	}
	if first.RemoteAddr != "10.0.0.9" {
		t.Errorf("RemoteAddr = %q, want the peer address", first.RemoteAddr)
	}
	if first.ClientIP == "1.2.3.4" {
		t.Error("a spoofed X-Forwarded-For was believed with no proxy trusted")
	}
	if first.UserAgent != "probe/1.0" {
		t.Errorf("UserAgent = %q", first.UserAgent)
	}
	if first.RequestID == "" {
		t.Error("every request must get an id")
	}
	// One request is one id, however many layers ask for the context.
	if second == nil || second.RequestID != first.RequestID {
		t.Error("the request id must be stable within a request")
	}
}

// A comma-separated query value and a repeated one decode to the same slice.
//
// parseQuery joins repeated parameters with commas and StringToSliceHookFunc
// splits them again, so a DTO declares []string and neither the handler nor the
// DTO has to take the string apart itself. Request DTOs rely on this; a change
// to either half would otherwise show up as levels quietly never matching.
func TestParseQuerySplitsListsEitherWayTheyAreWritten(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &BaseHandler{}

	type query struct {
		Levels []string `mapstructure:"levels"`
		Search string   `mapstructure:"search"`
		Limit  int      `mapstructure:"limit"`
	}

	for _, url := range []string{
		"/?levels=error,warn&search=boom&limit=500",
		"/?levels=error&levels=warn&search=boom&limit=500",
	} {
		var got query
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, url, nil)

		if err := h.parseQuery(ctx, &got); err != nil {
			t.Fatalf("parseQuery(%s): %v", url, err)
		}
		if len(got.Levels) != 2 || got.Levels[0] != "error" || got.Levels[1] != "warn" {
			t.Errorf("parseQuery(%s) levels = %#v, want [error warn]", url, got.Levels)
		}
		if got.Search != "boom" || got.Limit != 500 {
			t.Errorf("parseQuery(%s) = %#v, want search=boom limit=500", url, got)
		}
	}

	// A parameter that was not sent stays empty rather than becoming [""].
	var empty query
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/?search=boom", nil)
	if err := h.parseQuery(ctx, &empty); err != nil {
		t.Fatalf("parseQuery: %v", err)
	}
	if len(empty.Levels) != 0 {
		t.Errorf("absent levels = %#v, want empty", empty.Levels)
	}
}
