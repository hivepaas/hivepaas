// Package mcp serves the Model Context Protocol at <API base path>/mcp: tools an
// AI client calls as the person whose API key it holds, answered through the
// dashboard's own endpoints. See
// docs/superpowers/specs/2026-09-26-mcp-server-design.md.
package mcp

import (
	"context"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

// forwardedHeaders are what the dispatcher copies from the MCP request onto the
// requests it sends: what says where the call came from, so that the handlers'
// audit entries show the client rather than the backend itself.
var forwardedHeaders = []string{"User-Agent", "X-Forwarded-For", "X-Real-Ip", "Accept-Language"}

// caller is who sent an MCP request, with which key, and from where.
type caller struct {
	auth  *basedto.Auth
	keyID string
	// access is what this request may change: what the setting allows, and the
	// key's access actions.
	access     access
	remoteAddr string
	header     http.Header
}

type callerKey struct{}

// withCaller is the context the endpoint hands the SDK: the tools find their
// caller in it.
func withCaller(ctx context.Context, c *caller, r *http.Request) context.Context {
	header := http.Header{}
	for _, name := range forwardedHeaders {
		if values := r.Header.Values(name); len(values) > 0 {
			header[http.CanonicalHeaderKey(name)] = values
		}
	}
	c.remoteAddr, c.header = r.RemoteAddr, header
	return context.WithValue(ctx, callerKey{}, c)
}

func callerFrom(ctx context.Context) *caller {
	c, _ := ctx.Value(callerKey{}).(*caller)
	return c
}
