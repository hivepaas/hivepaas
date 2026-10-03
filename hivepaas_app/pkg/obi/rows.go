package obi

import (
	"encoding/json"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

// The rows the agent writes, by their "hp" field.
const (
	// RowRoutes is an app's requests to one route, method and status class.
	RowRoutes = "routes"
	// RowCalls is an app's calls to one peer: another app, a database, a host.
	RowCalls = "calls"
)

// The kinds of call a row is of.
const (
	KindHTTP = "http"
	KindDB   = "db"
	KindRPC  = "rpc"
)

// family is a histogram OBI exports, and the row it becomes.
type family struct {
	row  string
	kind string
}

// families are the histograms read: what an app serves, and what it calls.
var families = map[string]family{
	"http_server_request_duration_seconds": {row: RowRoutes, kind: KindHTTP},
	"rpc_server_duration_seconds":          {row: RowRoutes, kind: KindRPC},
	"http_client_request_duration_seconds": {row: RowCalls, kind: KindHTTP},
	"rpc_client_duration_seconds":          {row: RowCalls, kind: KindRPC},
	"db_client_operation_duration_seconds": {row: RowCalls, kind: KindDB},
}

// familyOf splits a sample's name into its histogram and part: _bucket, _sum
// or _count. "" when it is no histogram read.
func familyOf(name string) (string, string) {
	for _, part := range []string{"_bucket", "_sum", "_count"} {
		if base, ok := strings.CutSuffix(name, part); ok {
			if _, known := families[base]; known {
				return base, part
			}
		}
	}
	return "", ""
}

// series is one histogram's series as last read: cumulative, as OBI keeps it.
type series struct {
	family  string
	labels  map[string]string
	count   float64
	sum     float64
	buckets map[string]float64 // by le, in seconds as OBI writes it
}

// Deltas turns OBI's cumulative histograms into what moved between two
// scrapes. The first scrape is a baseline: what it holds happened before, at
// times unknown. A series first seen after it is new since the last scrape,
// and counts whole; one whose count went down was restarted, and counts from
// zero.
type Deltas struct {
	last   map[string]*series
	primed bool
}

// NewDeltas is a tracker with nothing read yet.
func NewDeltas() *Deltas {
	return &Deltas{last: map[string]*series{}}
}

// Reset forgets every series: the next scrape is a baseline again, as when
// OBI was replaced.
func (d *Deltas) Reset() {
	d.last = map[string]*series{}
	d.primed = false
}

// Row is one series' change over an interval, attributed to an app.
type Row struct {
	HP        string
	App       string
	Container string
	Kind      string
	Method    string
	Route     string
	Peer      string
	Operation string
	Status    string
	Count     int64
	Errors    int64
	SumMs     float64
	// Buckets are the requests at or under each bound, in milliseconds, as a
	// Prometheus histogram's are: cumulative over the bounds.
	Buckets map[float64]int64
}

// Read takes a scrape of OBI's metrics and answers the rows of what moved
// since the last, for the containers appOf names an app of; the others are
// dropped.
func (d *Deltas) Read(r io.Reader, appOf func(container string) (string, bool)) ([]*Row, error) {
	samples, err := parseText(r, func(name string) bool {
		f, _ := familyOf(name)
		return f != ""
	})
	if err != nil {
		return nil, err
	}
	now := collect(samples)
	var rows []*Row
	for key, cur := range now {
		prev := d.last[key]
		d.last[key] = cur
		var delta *series
		switch {
		case prev == nil && !d.primed:
			continue
		case prev == nil || cur.count < prev.count:
			delta = cur
		default:
			delta = minus(cur, prev)
		}
		if delta.count <= 0 {
			continue
		}
		app, ok := appOf(cur.labels["container_name"])
		if !ok {
			continue
		}
		rows = append(rows, rowOf(delta, app))
	}
	// A series gone from the scrape - its container gone - is forgotten.
	for key := range d.last {
		if _, ok := now[key]; !ok {
			delete(d.last, key)
		}
	}
	d.primed = true
	sort.Slice(rows, func(i, j int) bool { return rows[i].key() < rows[j].key() })
	return rows, nil
}

// collect gathers a scrape's samples into series, by family and labels.
func collect(samples []sample) map[string]*series {
	out := map[string]*series{}
	for _, s := range samples {
		fam, part := familyOf(s.Name)
		le := s.Labels["le"]
		delete(s.Labels, "le")
		key := seriesKey(fam, s.Labels)
		cur := out[key]
		if cur == nil {
			cur = &series{family: fam, labels: s.Labels, buckets: map[string]float64{}}
			out[key] = cur
		}
		switch part {
		case "_count":
			cur.count = s.Value
		case "_sum":
			cur.sum = s.Value
		case "_bucket":
			cur.buckets[le] = s.Value
		}
	}
	return out
}

func seriesKey(fam string, labels map[string]string) string {
	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString(fam)
	for _, name := range names {
		b.WriteString("|" + name + "=" + labels[name])
	}
	return b.String()
}

func minus(cur, prev *series) *series {
	out := &series{family: cur.family, labels: cur.labels, count: cur.count - prev.count, sum: cur.sum - prev.sum,
		buckets: make(map[string]float64, len(cur.buckets))}
	for le, v := range cur.buckets {
		out.buckets[le] = v - prev.buckets[le]
	}
	return out
}

// rowOf is a series' change as a row: what OBI's labels say of it, by family.
func rowOf(s *series, app string) *Row {
	f := families[s.family]
	l := s.labels
	r := &Row{HP: f.row, App: app, Container: l["container_name"], Kind: f.kind, Count: int64(math.Round(s.count)),
		SumMs: s.sum * 1000, Buckets: make(map[float64]int64, len(s.buckets))} //nolint:mnd // s to ms
	for le, v := range s.buckets {
		bound, err := parseValue(le)
		if err != nil {
			continue
		}
		r.Buckets[bound*1000] = int64(math.Round(v)) //nolint:mnd // s to ms
	}
	failed := l["error_type"] != ""
	switch f.kind {
	case KindHTTP:
		r.Method, r.Route = l["http_request_method"], l["http_route"]
		r.Status = statusClass(l["http_response_status_code"])
		failed = failed || r.Status == "5xx"
		if f.row == RowCalls {
			r.Peer = peerOf(l["server_address"], l["server_port"])
		}
	case KindRPC:
		r.Method, r.Status = l["rpc_method"], l["rpc_grpc_status_code"]
		failed = failed || (r.Status != "" && r.Status != "0")
		if f.row == RowCalls {
			r.Peer = peerOf(l["server_address"], l["server_port"])
		}
	case KindDB:
		// No address: OBI writes "outgoing". The database is its system and
		// name, which the API matches to an app.
		r.Peer = strings.Trim(l["db_system_name"]+"/"+l["db_namespace"], "/")
		r.Operation = l["db_operation_name"]
		failed = failed || l["db_response_status_code"] != ""
	}
	if failed {
		r.Errors = r.Count
	}
	return r
}

func statusClass(code string) string {
	if len(code) != 3 || code[0] < '1' || code[0] > '5' { //nolint:mnd // an HTTP status
		return ""
	}
	return code[:1] + "xx"
}

func peerOf(address, port string) string {
	if address == "" || address == "outgoing" {
		return ""
	}
	if port == "" {
		return address
	}
	return address + ":" + port
}

func (r *Row) key() string {
	return r.HP + "|" + r.App + "|" + r.Container + "|" + r.Kind + "|" + r.Peer + "|" + r.Method + "|" + r.Route +
		"|" + r.Operation + "|" + r.Status
}

// MarshalJSON writes a row as the agent's other rows are written: flat, its
// buckets as le<bound in ms> fields, "leInf" for the last.
func (r *Row) MarshalJSON() ([]byte, error) {
	m := map[string]any{"hp": r.HP, "app": r.App, "container": r.Container, "kind": r.Kind,
		"count": r.Count, "errors": r.Errors, "sumMs": math.Round(r.SumMs*1000) / 1000} //nolint:mnd // µs
	for name, v := range map[string]string{"method": r.Method, "route": r.Route, "peer": r.Peer,
		"operation": r.Operation, "status": r.Status} {
		if v != "" {
			m[name] = v
		}
	}
	for bound, n := range r.Buckets {
		m[BucketField(bound)] = n
	}
	return json.Marshal(m) //nolint:wrapcheck // a map of plain values
}

// BucketField is the field a bound's count is written in: le5 for 5 ms, le7.5
// for 7.5 ms, leInf for the last.
func BucketField(boundMs float64) string {
	if math.IsInf(boundMs, 1) {
		return "leInf"
	}
	return "le" + strconv.FormatFloat(boundMs, 'f', -1, 64)
}

// WriteRows writes rows one JSON line each.
func WriteRows(w io.Writer, rows []*Row) error {
	for _, r := range rows {
		line, err := json.Marshal(r)
		if err != nil {
			return err //nolint:wrapcheck // a map of plain values
		}
		if _, err = w.Write(append(line, '\n')); err != nil {
			return err //nolint:wrapcheck // the agent's stdout
		}
	}
	return nil
}
