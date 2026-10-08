package authhandler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// The demo user is a public account: it reads, and does nothing else.
//
// Its session is limited to READ, but that limit only holds where a handler
// asks for an access check, and asks for the right one. Here it holds for every
// request, by its method, wherever it is authenticated - from a header or
// dispatched by an MCP tool - before any handler can fall back on the caller's
// own user.

// demoWrites are the requests other than reads the demo user may make, by
// method and route suffix: they change nothing a visitor shares.
var demoWrites = []string{
	http.MethodPost + " /sessions/refresh",
	http.MethodDelete + " /sessions",
	http.MethodPost + " /sched-jobs/calc-next-runs",
}

// demoRefusedReads are the reads the demo user may not make, by route suffix:
// they act - a shell in a container, a GitHub App made - or hand over what a
// public demo must keep: a container's files, secrets, certificates' keys,
// backups.
var demoRefusedReads = []string{
	"/apps/:appID/terminal",
	"/apps/:appID/container/file-download",
	"/apps/:appID/container/file-upload/stream",
	"/secrets/:itemID/download",
	"/secrets/:itemID/download-token",
	"/ssl-certs/:itemID/download",
	"/backup-snapshots/:itemID/download",
	"/github-apps/:itemID/manifest-flow/begin",
	"/github-apps/:itemID/manifest-flow/progress",
}

// DemoAllows says whether the demo user may make a request, by its method and
// the route it matched (gin's FullPath).
func DemoAllows(method, route string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		for _, suffix := range demoRefusedReads {
			if strings.HasSuffix(route, suffix) {
				return false
			}
		}
		return true
	default:
		for _, write := range demoWrites {
			allowedMethod, suffix, _ := strings.Cut(write, " ")
			if method == allowedMethod && strings.HasSuffix(route, suffix) {
				return true
			}
		}
		return false
	}
}

// refuseDemo refuses a request of the demo user that DemoAllows does not allow.
func refuseDemo(ctx *gin.Context, user *basedto.User) error {
	if user == nil || user.User == nil || !user.IsDemoUser() {
		return nil
	}
	if DemoAllows(ctx.Request.Method, ctx.FullPath()) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
}
