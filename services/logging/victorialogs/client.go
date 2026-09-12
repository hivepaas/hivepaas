package victorialogs

import (
	"context"
	"fmt"
	"net/http"
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

	resp, err := http.DefaultClient.Do(req)
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

// Query is the read path. Building LogsQL safely is its own subject and is not
// part of collection; this exists so that Client satisfies loggingmodel.Backend.
func (c *Client) Query(_ context.Context, _ *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error) {
	return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).
		WithExtraDetail("the read path is not implemented yet")
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
