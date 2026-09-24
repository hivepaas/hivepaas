package dockerproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	// daemonHost is the host requests to the daemon carry. The transport dials the
	// daemon's socket whatever the address, so it only has to be a valid name.
	daemonHost = "docker"

	contentTypeJSON = "application/json"
)

// errDaemon is the daemon answering a lookup in a way the proxy cannot use.
var errDaemon = errors.New("unexpected answer from the Docker daemon")

// daemon asks the Docker daemon what the proxy needs to know to judge a request.
type daemon struct {
	client *http.Client
}

// get fetches path and decodes the answer into out. Not found is not an error:
// the bool says whether there was anything.
func (d *daemon) get(ctx context.Context, path string, out any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+daemonHost+path, nil)
	if err != nil {
		return false, fmt.Errorf("GET %s: %w", path, err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("GET %s: %w", path, err)
	}
	// What is left of the body is read before closing it, so that the connection
	// goes back to the pool: a decoder stops at the end of the value, before the
	// newline the daemon writes after it.
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
			return false, fmt.Errorf("GET %s: %w", path, err)
		}
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("%w: GET %s answered %d", errDaemon, path, resp.StatusCode)
	}
}

// labelFilter is a filters query parameter selecting by one label.
func labelFilter(key, value string) string {
	raw, _ := json.Marshal(map[string][]string{"label": {key + "=" + value}})
	return url.QueryEscape(string(raw))
}

// maxErrorDetail is how much of a refusal from the daemon is kept to explain it.
const maxErrorDetail = 512

// createVolume creates a volume with labels, and fails unless the daemon
// accepts it.
func (d *daemon) createVolume(ctx context.Context, name string, labels map[string]string) error {
	const path = "/volumes/create"
	raw, err := json.Marshal(map[string]any{fieldName: name, fieldLabels: labels})
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+daemonHost+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorDetail))
		return fmt.Errorf("%w: POST %s answered %d: %s", errDaemon, path, resp.StatusCode, bytes.TrimSpace(detail))
	}
	return nil
}
