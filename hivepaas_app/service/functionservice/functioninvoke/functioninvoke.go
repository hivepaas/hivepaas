// Package functioninvoke is the runtime's invoke as HivePaaS calls it: the
// request it reads on its standard input, and the result it writes as the last
// line of its standard output, after the call's log. A test run and a
// function-invoke job both call it. See CONTRACT.md of hivepaas/function-runtimes.
package functioninvoke

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	// ResultMarker starts the last line invoke writes: its result.
	ResultMarker = "#hivepaas-result "

	// lineMax is the longest piece of a log line held before it is passed on.
	lineMax = 64 * 1024
	// resultMax is the longest result line read: its body is in it, in base64.
	resultMax = 16 * int(unit.MB)
)

// Command is the runtime's invoke, on the PATH of every runtime image.
var Command = []string{"hivepaas-runtime", "invoke"}

// Request is a request as invoke reads it. Every field may be left out: the
// method is then GET, the path /, the body empty.
type Request struct {
	Method  string              `json:"method,omitempty"`
	Path    string              `json:"path,omitempty"`
	Query   map[string][]string `json:"query,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	// Body is in base64 on the wire.
	Body []byte `json:"body,omitempty"`
}

// RequestOf is the request a function-invoke job sends: its path's query read
// apart, its text body as bytes.
func RequestOf(invoke *entity.SchedJobFunctionInvoke) *Request {
	req := &Request{Method: invoke.Method, Path: invoke.Path, Headers: invoke.Headers}
	if path, rawQuery, found := strings.Cut(invoke.Path, "?"); found {
		req.Path = path
		// A query that does not parse keeps what parsed of it.
		query, _ := url.ParseQuery(rawQuery)
		if len(query) > 0 {
			req.Query = query
		}
	}
	if invoke.Body != "" {
		req.Body = []byte(invoke.Body)
	}
	return req
}

// Result is what invoke wrote of the call: the handler's response, and how
// the call went.
type Result struct {
	Status     int                 `json:"status"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	RequestID  string              `json:"requestId"`
	DurationMs float64             `json:"durationMs"`
	// Outcome is ok, error or timeout.
	Outcome string `json:"outcome"`
}

// ParseResult reads a result line, its marker taken off.
func ParseResult(line string) (*Result, error) {
	result := &Result{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), result); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return result, nil
}

// OutputWriter takes invoke's standard output as it comes: each line of the
// call's log goes on to onLog, and the result line is kept. A log line longer
// than lineMax goes on in pieces.
type OutputWriter struct {
	onLog     func(line []byte)
	line      bytes.Buffer
	result    []byte
	tooLarge  bool
	resultMax int
	// dropping is a result line too large, whose rest is dropped up to its end.
	dropping bool
}

func NewOutputWriter(onLog func(line []byte)) *OutputWriter {
	return &OutputWriter{onLog: onLog, resultMax: resultMax}
}

func (w *OutputWriter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			w.line.Write(p)
			w.passOnLongLine()
			break
		}
		w.line.Write(p[:i+1])
		w.endLine()
		p = p[i+1:]
	}
	return n, nil
}

// Flush ends the line written last, which may lack its newline.
func (w *OutputWriter) Flush() {
	if w.line.Len() > 0 {
		w.endLine()
	}
}

// Result is the result line, its marker taken off; nil when invoke wrote none.
func (w *OutputWriter) Result() []byte {
	return w.result
}

// ResultTooLarge says that invoke wrote a result too large to be read.
func (w *OutputWriter) ResultTooLarge() bool {
	return w.tooLarge
}

func (w *OutputWriter) endLine() {
	line := w.line.Bytes()
	switch {
	case w.dropping:
		w.dropping = false
	case bytes.HasPrefix(line, []byte(ResultMarker)):
		if len(line) > w.resultMax {
			w.result, w.tooLarge = nil, true
		} else {
			w.result = bytes.TrimSpace(bytes.Clone(line[len(ResultMarker):]))
			w.tooLarge = false
		}
	default:
		w.onLog(line)
	}
	w.line.Reset()
}

// passOnLongLine passes on a log line too long to hold. A line that may be the
// result is held up to the result's limit, and then dropped as too large.
func (w *OutputWriter) passOnLongLine() {
	line := w.line.Bytes()
	marker := []byte(ResultMarker)
	mayBeResult := bytes.HasPrefix(line, marker) || bytes.HasPrefix(marker, line)
	switch {
	case w.dropping:
		w.line.Reset()
	case mayBeResult && len(line) > w.resultMax:
		w.result, w.tooLarge, w.dropping = nil, true, true
		w.line.Reset()
	case !mayBeResult && len(line) > lineMax:
		w.onLog(line)
		w.line.Reset()
	}
}
