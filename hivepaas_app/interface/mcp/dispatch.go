package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
)

// maxDispatchedBody bounds what the dispatcher keeps of an answer. Tools cut
// their own output far below it; this is what stops a handler that answers
// more than expected from costing the backend its memory.
const maxDispatchedBody = 16 << 20

var (
	// errNoCaller is a tool running without the caller the endpoint puts in the
	// context: a bug, never a person's mistake.
	errNoCaller = errors.New("mcp: no caller in the context")
	// errNotDispatchable is a path a tool may not send a request to.
	errNotDispatchable = errors.New("mcp: not a path a tool may dispatch to")
	// errAnswerTooLarge is a handler answering more than maxDispatchedBody.
	errAnswerTooLarge = errors.New("mcp: the answer is too large")
)

// Dispatcher sends a request to the backend's own router, as the caller: the
// request the dashboard would send, answered by the same handler with the same
// checks. It never goes through the network.
type Dispatcher struct {
	handler  http.Handler
	basePath string
}

func NewDispatcher(handler http.Handler, basePath string) *Dispatcher {
	return &Dispatcher{handler: handler, basePath: strings.TrimSuffix(basePath, "/")}
}

// Response is what the handler answered, when it answered 2xx.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// APIError is a handler's non-2xx answer, as the API words it.
type APIError struct {
	Status  int
	Code    string
	Title   string
	Detail  string
	Details []string
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d", e.Status)
	for _, part := range []string{e.Title, e.Detail} {
		if part != "" {
			b.WriteString(": " + part)
		}
	}
	for _, detail := range e.Details {
		b.WriteString("; " + detail)
	}
	return b.String()
}

// Do dispatches one request as the caller in ctx. path is below the API's base
// path, its segments already escaped; query and body may be nil. A path under
// /mcp is refused: the endpoint does not call itself.
func (d *Dispatcher) Do(ctx context.Context, method, path string, query url.Values, body any) (*Response, error) {
	c := callerFrom(ctx)
	if c == nil {
		return nil, errNoCaller
	}
	if path == "/mcp" || strings.HasPrefix(path, "/mcp/") || !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("%w: %s", errNotDispatchable, path)
	}

	target := d.basePath + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(authhandler.WithDispatchedAuth(ctx, c.auth), method, target, reader)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for name, values := range c.header {
		req.Header[name] = values
	}
	req.RemoteAddr = c.remoteAddr

	w := &bufferedResponse{header: http.Header{}}
	d.handler.ServeHTTP(w, req)
	if w.overflow {
		return nil, fmt.Errorf("%w: %s %s answered more than %d bytes", errAnswerTooLarge, method, path,
			maxDispatchedBody)
	}
	status := w.statusCode()
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, apiErrorOf(status, w.body.Bytes())
	}
	return &Response{Status: status, Header: w.header, Body: w.body.Bytes()}, nil
}

// apiErrorOf reads the API's error body. The fields that exist for debugging -
// cause, debug log, stack trace - are left out: they are for a developer at the
// backend, not for a model.
func apiErrorOf(status int, body []byte) *APIError {
	var info struct {
		Title  string `json:"title"`
		Code   string `json:"code"`
		Detail string `json:"detail"`
		Errors []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	apiErr := &APIError{Status: status}
	if json.Unmarshal(body, &info) != nil {
		apiErr.Title = http.StatusText(status)
		return apiErr
	}
	apiErr.Title, apiErr.Code, apiErr.Detail = info.Title, info.Code, info.Detail
	for _, inner := range info.Errors {
		detail := inner.Message
		if inner.Path != "" {
			detail = inner.Path + ": " + inner.Message
		}
		apiErr.Details = append(apiErr.Details, detail)
	}
	return apiErr
}

// bufferedResponse is what a dispatched request is answered into.
type bufferedResponse struct {
	header   http.Header
	status   int
	body     bytes.Buffer
	overflow bool
}

func (w *bufferedResponse) Header() http.Header { return w.header }

func (w *bufferedResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *bufferedResponse) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.body.Len()+len(p) > maxDispatchedBody {
		w.overflow = true
		return 0, errAnswerTooLarge
	}
	// A bytes.Buffer grows or panics; it never returns an error.
	_, _ = w.body.Write(p)
	return len(p), nil
}

// Flush is a no-op: nothing is streamed to anybody.
func (w *bufferedResponse) Flush() {}

func (w *bufferedResponse) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
