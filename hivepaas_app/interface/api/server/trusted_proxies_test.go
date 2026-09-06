package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// clientIPWith reports what gin resolves the caller's address to when the caller
// volunteers someone else's address in X-Forwarded-For.
func clientIPWith(t *testing.T, configure func(*gin.Engine)) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	configure(engine)

	got := ""
	engine.GET("/", func(c *gin.Context) { got = c.ClientIP() })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:5555"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	engine.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestApplyTrustedProxies(t *testing.T) {
	// What gin does on its own, and the reason this code exists: the caller picks
	// the address that would be recorded against them.
	if got := clientIPWith(t, func(*gin.Engine) {}); got != "1.2.3.4" {
		t.Fatalf("gin's default changed; got %q - revisit applyTrustedProxies", got)
	}

	tests := []struct {
		name    string
		trusted []string
		want    string
	}{
		{
			name: "nothing configured means nothing trusted",
			want: "10.0.0.9",
		},
		{
			name:    "a malformed entry falls back to trusting nothing",
			trusted: []string{"not-a-cidr"},
			want:    "10.0.0.9",
		},
		{
			name:    "one bad entry does not leave the good ones trusted",
			trusted: []string{"10.0.0.0/8", "not-a-cidr"},
			want:    "10.0.0.9",
		},
		{
			name:    "a configured proxy is believed",
			trusted: []string{"10.0.0.0/8"},
			want:    "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clientIPWith(t, func(e *gin.Engine) {
				applyTrustedProxies(e, tt.trusted, nil)
			})
			if got != tt.want {
				t.Errorf("ClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
