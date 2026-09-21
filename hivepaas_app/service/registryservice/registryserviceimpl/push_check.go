package registryserviceimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

const (
	// defaultPushCheckBytes is above the 100 MB a proxy on a free plan allows per
	// request, which is the limit worth finding out about before a build does.
	defaultPushCheckBytes = 150 << 20
	maxPushCheckBytes     = 1 << 30

	// pushCheckRepo is thrown away: the upload is canceled, and whatever it
	// leaves behind is an unfinished upload, which zot's garbage collector
	// removes on its next pass.
	pushCheckRepo    = "hivepaas-selftest"
	pushCheckTimeout = 10 * time.Minute
)

// CheckPush sends a large blob through the public domain and throws it away.
//
// It is the only answer that is not a guess: headers say what is in front of the
// registry, but only an upload says whether that thing will carry a layer.
func (s *service) CheckPush(
	ctx context.Context,
	db database.IDB,
	req *registryservice.PushCheckReq,
) (*registryservice.PushCheckResult, error) {
	setting := req.Setting
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}
	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled || cfg.Domain == "" {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}

	size := req.Bytes
	if size <= 0 {
		size = defaultPushCheckBytes
	}
	if size > maxPushCheckBytes {
		size = maxPushCheckBytes
	}

	user, password, err := s.credentialOf(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	client := &http.Client{Timeout: pushCheckTimeout}
	started := time.Now()

	location, status, err := s.startUpload(ctx, client, cfg.Domain, user, password)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if location == "" {
		result := readPushCheckAnswer(status)
		result.Elapsed = time.Since(started)
		return result, nil
	}

	status, err = s.sendBlob(ctx, client, location, user, password, size)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Canceling leaves an unfinished upload rather than a repository, whatever
	// the answer was.
	s.cancelUpload(ctx, client, location, user, password)

	result := readPushCheckAnswer(status)
	result.Elapsed = time.Since(started)
	return result, nil
}

func (s *service) startUpload(ctx context.Context, client *http.Client, domain, user, password string) (
	string, int, error) {
	url := fmt.Sprintf("https://%s/v2/%s/blobs/uploads/", domain, pushCheckRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", 0, hperrors.Wrap(err)
	}
	req.SetBasicAuth(user, password)

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		return "", resp.StatusCode, nil
	}
	return absoluteLocation(domain, resp.Header.Get("Location")), resp.StatusCode, nil
}

func (s *service) sendBlob(
	ctx context.Context,
	client *http.Client,
	location, user, password string,
	size int64,
) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, location,
		io.LimitReader(zeroReader{}, size))
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	req.SetBasicAuth(user, password)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size

	resp, err := client.Do(req)
	if err != nil {
		return 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

func (s *service) cancelUpload(ctx context.Context, client *http.Client, location, user, password string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, location, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(user, password)

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Warn("could not cancel the registry push check upload", "error", err.Error())
		return
	}
	_ = resp.Body.Close()
}

// absoluteLocation turns the Location a registry hands back into a URL. The
// specification allows a relative one, and most registries send exactly that.
func absoluteLocation(domain, location string) string {
	if location == "" {
		return ""
	}
	if len(location) > 4 && (location[:5] == "https" || location[:4] == "http") {
		return location
	}
	return fmt.Sprintf("https://%s%s", domain, location)
}

// readPushCheckAnswer turns a status code into the sentence the dashboard shows.
func readPushCheckAnswer(status int) *registryservice.PushCheckResult {
	switch status {
	case http.StatusAccepted, http.StatusCreated, http.StatusNoContent:
		return &registryservice.PushCheckResult{
			OK: true, StatusCode: status,
			Detail: "The upload went through: nothing in front of the registry is limiting request bodies.",
		}
	case http.StatusRequestEntityTooLarge:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail: "Something in front of the registry refused the upload with a body limit. " +
				"Set this domain to DNS-only so that registry traffic reaches the cluster directly.",
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail:     "The registry refused the credential. Rotating it and redeploying may be needed.",
		}
	default:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail:     fmt.Sprintf("The upload was answered with %d.", status),
		}
	}
}

// zeroReader is the body of the check: bytes nobody has to hold in memory.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
