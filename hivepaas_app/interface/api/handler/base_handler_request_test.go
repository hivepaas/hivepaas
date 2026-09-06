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
