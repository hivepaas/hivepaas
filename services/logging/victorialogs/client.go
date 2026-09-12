package victorialogs

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const pingTimeout = 5 * time.Second

// Ping reports whether the backend is serving.
func (c *Client) Ping(ctx context.Context) error {
	if c.cfg.Endpoint.URL == "" {
		return hperrors.Wrap(loggingmodel.ErrIngestEndpointRequired)
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimSuffix(c.cfg.Endpoint.URL, "/")+HealthPath, nil)
	if err != nil {
		return hperrors.Wrap(err)
	}
	applyAuth(req, &c.cfg.Endpoint)

	resp, err := c.httpClient(pingTimeout).Do(req)
	if err != nil {
		return hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return hperrors.Wrap(loggingmodel.ErrBackendUnreachable).
			WithExtraDetail("health returned %d", resp.StatusCode)
	}
	return nil
}

const (
	queryTimeout = 30 * time.Second
	// maxLineBytes bounds one returned line. A log line longer than this is
	// cut by the scanner and reported as an error rather than read unbounded.
	maxLineBytes = 1 << 20
	// errorBodyBytes is how much of an error response is kept for the detail.
	errorBodyBytes = 512
)

// Query runs a search. The request is turned into LogsQL by BuildQuery, never
// passed through.
func (c *Client) Query(ctx context.Context, req *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error) {
	q, err := BuildQuery(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if c.cfg.Endpoint.URL == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("no query endpoint")
	}

	form := url.Values{"query": {q}}
	if !req.Start.IsZero() {
		form.Set("start", req.Start.UTC().Format(time.RFC3339Nano))
	}
	if !req.End.IsZero() {
		form.Set("end", req.End.UTC().Format(time.RFC3339Nano))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(c.cfg.Endpoint.URL, "/")+QueryPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyAuth(httpReq, &c.cfg.Endpoint)

	resp, err := c.httpClient(queryTimeout).Do(httpReq)
	if err != nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyBytes))
		base := loggingmodel.ErrBackendUnreachable
		if resp.StatusCode == http.StatusBadRequest {
			base = loggingmodel.ErrQueryInvalid
		}
		return nil, hperrors.Wrap(base).WithExtraDetail("%d: %s", resp.StatusCode, string(body))
	}

	entries, err := readEntries(resp.Body)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &loggingmodel.QueryResp{Entries: entries, Truncated: len(entries) == req.Limit}, nil
}

// readEntries parses one JSON object per line, newest first as the query sorts
// them, and returns them oldest first - the order a person reads.
func readEntries(r io.Reader) ([]loggingmodel.LogEntry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxLineBytes) //nolint:mnd // initial buffer
	var out []loggingmodel.LogEntry
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var row map[string]string
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("unreadable line: %s", err.Error())
		}
		ts, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("unreadable _time: %s", err.Error())
		}
		out = append(out, loggingmodel.LogEntry{
			Time:    ts,
			Message: row["_msg"],
			Stream:  row["stream"],
			Level:   row[LevelField],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	slices.Reverse(out)
	return out, nil
}

// httpClient honors the endpoint's TLS setting. http.DefaultClient would
// silently ignore TLSSkipVerify.
func (c *Client) httpClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // stdlib guarantees it
	if c.cfg.Endpoint.TLSSkipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // the operator asked for it
	}
	return &http.Client{Transport: tr, Timeout: timeout}
}

func applyAuth(req *http.Request, ep *loggingmodel.Endpoint) {
	if ep.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+ep.BearerToken)
	} else if ep.Username != "" {
		req.SetBasicAuth(ep.Username, ep.Password)
	}
	for k, v := range ep.Headers {
		req.Header.Set(k, v)
	}
}

// IngestURL is where a collector should write, derived from the endpoint.
func (c *Client) IngestURL() string {
	return fmt.Sprintf("%s%s", strings.TrimSuffix(c.cfg.Endpoint.URL, "/"), IngestPath)
}
