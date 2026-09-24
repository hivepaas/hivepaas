package dockerproxy

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"path"
	"regexp"
	"slices"
	"sync/atomic"
)

// flushImmediately makes the forwarder write each chunk as it arrives: logs,
// attach and pull progress are streams.
const flushImmediately = -1

// versionPrefix is the API version a client may put in front of any path.
var versionPrefix = regexp.MustCompile(`^/v[0-9]+\.[0-9]+`)

// Decision is what the proxy did with one request.
type Decision struct {
	AppID   string
	Method  string
	Path    string
	Allowed bool
	// Reason is the rule that let the request through, or the one it broke.
	Reason string
}

// Options are what a proxy needs besides its policy.
type Options struct {
	// Upstream reaches the Docker daemon. Requests are sent to http://docker/...,
	// so it is a transport that dials the daemon's socket whatever the address.
	Upstream http.RoundTripper
	// OnDecision, when set, is told about every request.
	OnDecision func(Decision)
}

// Proxy serves one app's socket.
type Proxy struct {
	policy     atomic.Pointer[Policy]
	daemon     *daemon
	forwarder  *httputil.ReverseProxy
	onDecision func(Decision)
}

// New returns a proxy enforcing policy.
func New(policy *Policy, opts Options) *Proxy {
	p := &Proxy{
		daemon:     &daemon{client: &http.Client{Transport: opts.Upstream}},
		onDecision: opts.OnDecision,
	}
	p.policy.Store(policy)
	p.forwarder = &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL.Scheme, r.Out.URL.Host, r.Out.Host = "http", daemonHost, daemonHost
		},
		Transport:     opts.Upstream,
		FlushInterval: flushImmediately,
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadGateway, "hivepaas: the Docker daemon did not answer: "+err.Error())
		},
	}
	return p
}

// SetPolicy replaces the policy. The next request is judged by the new one; a
// request in flight finishes under the old.
func (p *Proxy) SetPolicy(policy *Policy) {
	p.policy.Store(policy)
}

// call is one request being judged.
type call struct {
	w      http.ResponseWriter
	r      *http.Request
	policy *Policy
	// args are the parts of the path the route captured: an id, a name.
	args []string
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &call{w: w, r: r, policy: p.policy.Load()}
	endpoint := versionPrefix.ReplaceAllString(r.URL.Path, "")
	// A path the proxy could read one way and the daemon another would be judged
	// by the wrong rule, so only the plain form is taken.
	if r.URL.RawPath != "" || endpoint != path.Clean(endpoint) {
		p.refuse(c, refusef("the path %s is not in plain form", r.URL.Path))
		return
	}
	for _, rt := range routes {
		if !slices.Contains(rt.methods, r.Method) {
			continue
		}
		match := rt.pattern.FindStringSubmatch(endpoint)
		if match == nil {
			continue
		}
		if !c.policy.allows(rt.group) {
			p.refuse(c, refusef("%s is not allowed for this app", rt.group))
			return
		}
		c.args = match[1:]
		rt.handle(p, c)
		return
	}
	p.refuse(c, refusef("%s %s is not an endpoint this app may use", r.Method, endpoint))
}

func (p *Proxy) forward(c *call, reason string) {
	p.decide(c, true, reason)
	p.forwarder.ServeHTTP(c.w, c.r)
}

// refuse answers the way the daemon answers a request it will not do, so that
// the app reports it the way it reports any daemon error.
func (p *Proxy) refuse(c *call, err error) {
	status := http.StatusBadGateway
	var refusal *refusalError
	if errors.As(err, &refusal) {
		status = http.StatusForbidden
	}
	p.decide(c, false, err.Error())
	writeError(c.w, status, "hivepaas: "+err.Error())
}

func (p *Proxy) decide(c *call, allowed bool, reason string) {
	if p.onDecision == nil {
		return
	}
	p.onDecision(Decision{
		AppID: c.policy.AppID, Method: c.r.Method, Path: c.r.URL.Path, Allowed: allowed, Reason: reason,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
