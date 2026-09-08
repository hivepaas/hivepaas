// Package reqinfo carries the few facts about an HTTP request that the layers
// below the handler need - who called, and from where - without those layers
// having to know about gin, or about HTTP at all.
//
// It exists for the audit log: an entry that cannot say where an action came
// from is worth much less than one that can.
package reqinfo

import (
	"context"
)

type ctxKey struct{}

// RequestInfo is what a record needs to attribute an action to a caller.
type RequestInfo struct {
	// RequestID ties together everything written for one request.
	RequestID string

	// ClientIP is the address after the proxy headers have been applied. It is
	// only as trustworthy as the trusted-proxy configuration: with no proxy
	// trusted it equals RemoteAddr, and with the wrong ones trusted a caller can
	// pick it by sending an X-Forwarded-For header of their choosing.
	ClientIP string

	// RemoteAddr is the address of the TCP peer, without the port. Nothing the
	// client sends can change it, which is why it is recorded next to ClientIP
	// rather than instead of it: if the proxy configuration is ever wrong, this
	// is the field that still means something.
	RemoteAddr string

	UserAgent string
}

// NewContext returns a context carrying info.
func NewContext(ctx context.Context, info *RequestInfo) context.Context {
	if info == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, info)
}

// From returns the request info carried by ctx, or nil when there is none.
//
// Callers must cope with nil rather than assume it: plenty of work runs outside
// any request - workers, scheduled jobs, startup, the agent.
func From(ctx context.Context) *RequestInfo {
	if ctx == nil {
		return nil
	}
	info, _ := ctx.Value(ctxKey{}).(*RequestInfo)
	return info
}

// ClientIPFrom is the address the request came from.
func ClientIPFrom(ctx context.Context) string {
	info := From(ctx)
	if info == nil {
		return ""
	}
	return info.ClientIP
}
