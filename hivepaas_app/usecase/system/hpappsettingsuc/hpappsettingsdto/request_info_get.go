package hpappsettingsdto

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

// proxyHeadersToReport are the headers this endpoint echoes back.
//
// An allowlist, not a dump of everything that arrived. The point of the endpoint
// is to show how a request reached the app, and reflecting arbitrary headers into
// a response would reflect the caller's own cookies and authorization along with
// it - a needless way to put credentials somewhere they can be copied out of.
var proxyHeadersToReport = []string{
	"X-Forwarded-For",
	"X-Forwarded-Proto",
	"X-Forwarded-Host",
	"X-Forwarded-Port",
	"X-Real-Ip",
	"Forwarded",
	"CF-Connecting-IP",
	"CF-Ray",
	"True-Client-IP",
}

// CollectProxyHeaders picks the reportable headers out of a request. It takes the
// lookup rather than the request so the analysis stays testable without one.
func CollectProxyHeaders(get func(string) string) map[string]string {
	headers := make(map[string]string, len(proxyHeadersToReport))
	for _, name := range proxyHeadersToReport {
		if value := get(name); value != "" {
			headers[name] = value
		}
	}
	return headers
}

type GetRequestInfoResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *RequestInfoResp `json:"data"`
}

type RequestInfoResp struct {
	// RemoteAddr is the peer the app is talking to - the proxy, or Traefik.
	RemoteAddr string `json:"remoteAddr"`
	// ClientIP is what the app currently believes the caller's address is.
	ClientIP string `json:"clientIp"`

	// ForwardedFor is the X-Forwarded-For chain as the app receives it, in order.
	ForwardedFor []string `json:"forwardedFor"`
	// Headers are the proxy-related headers that arrived, verbatim.
	Headers map[string]string `json:"headers"`

	// SuggestedProxyHops is the value to put in the proxy settings, derived from
	// the chain below. Read Explanation before trusting it.
	SuggestedProxyHops int    `json:"suggestedProxyHops"`
	Explanation        string `json:"explanation"`
}

type RequestInfoTransformInput struct {
	RemoteAddr   string
	ClientIP     string
	HeaderValues map[string]string
}

// TransformRequestInfo describes how the request reached the app, and what
// forwarded-header depth that implies.
//
// The depth cannot be read off the deployment diagram: whether a proxy adds
// itself to X-Forwarded-For is that proxy's decision, and the number of entries
// is the only thing that settles it. So it is measured here rather than assumed.
//
// The arithmetic: what a Traefik middleware sees is what arrived at Traefik, and
// what the app sees is that plus the one entry Traefik appends when it forwards.
// So the chain the middleware worked from is one shorter than the chain below,
// and the caller sits at its far left - which is a depth of len(chain)-1 counting
// from the right.
func TransformRequestInfo(input *RequestInfoTransformInput) *RequestInfoResp {
	forwardedFor := parseForwardedFor(input.HeaderValues["X-Forwarded-For"])

	resp := &RequestInfoResp{
		RemoteAddr:   input.RemoteAddr,
		ClientIP:     input.ClientIP,
		ForwardedFor: forwardedFor,
		Headers:      input.HeaderValues,
	}

	if len(forwardedFor) > 1 {
		resp.SuggestedProxyHops = len(forwardedFor) - 1
	}
	resp.Explanation = explainRequestInfo(forwardedFor)

	return resp
}

func parseForwardedFor(header string) []string {
	if strings.TrimSpace(header) == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	res := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

// explainRequestInfo says what to do with the numbers, including how to tell that
// the measurement itself was taken wrongly.
func explainRequestInfo(forwardedFor []string) string {
	const check = " Call this endpoint from outside the cluster, over the same path real " +
		"traffic takes, and check that the first entry of forwardedFor is your own public " +
		"address - if it is not, the reading is of the wrong path and the suggestion is wrong."

	switch {
	case len(forwardedFor) == 0:
		return "No X-Forwarded-For arrived. Either nothing is proxying this request, or a " +
			"proxy in front is not adding one. Leave proxyHops unset." + check
	case len(forwardedFor) == 1:
		return "Only one entry, which is the address Traefik itself added when forwarding. " +
			"Nothing is in front of Traefik on this path. Leave proxyHops unset." + check
	default:
		return "Set proxyHops to suggestedProxyHops, and make sure trustedIPs covers the " +
			"addresses of every proxy in this chain except the first entry, which is the " +
			"caller." + check
	}
}
