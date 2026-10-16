package licenseserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

// The license portal is HivePaaS's own, not a system of someone else's, so its
// client lives here rather than in services/ (docs/ARCHITECTURE.md §6).
const (
	portalTimeout = 30 * time.Second
	// portalUserAgent names HivePaaS and not its version: a refresh sends the key
	// and the installation's ID, and nothing else.
	portalUserAgent = "HivePaaS"
	// maxPortalAnswerLen bounds what is read of an answer: a key, and a little
	// JSON around it.
	maxPortalAnswerLen = license.MaxKeyTextLen + 4<<10
	// maxPortalDetailLen bounds what of a refusal's detail is kept: it is shown
	// on System > License as the last check's error.
	maxPortalDetailLen = 300
)

// newPortalClient is the client refreshes ask the portal with. It follows no
// redirect: the key goes to the configured portal and nowhere else, and a portal
// that moved is told by its answer.
func newPortalClient() *http.Client {
	return &http.Client{
		Timeout: portalTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// portalRefreshReq is all a refresh sends.
type portalRefreshReq struct {
	Key            string `json:"key"`
	InstallationID string `json:"installationId"`
}

type portalRefreshResp struct {
	Data *portalRefreshData `json:"data"`
}

type portalRefreshData struct {
	Result  base.LicenseRefreshResult `json:"result"`
	Key     string                    `json:"key"`
	MovedAt *time.Time                `json:"movedAt"`
}

// portalError is what of the portal's error answer is read.
type portalError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// portalAnswer is the portal's answer: newer with the key, moved with when, or
// current; unknown for a key it did not issue; failed, with why, for none.
type portalAnswer struct {
	result  base.LicenseRefreshResult
	newKey  string
	movedAt time.Time
	err     error
}

// ask is the portal's answer about the key, for this installation.
func (s *service) ask(ctx context.Context, cfg *config.Config, keyText string) *portalAnswer {
	installationID, err := s.InstallationID(ctx)
	if err != nil {
		return &portalAnswer{result: base.LicenseRefreshFailed, err: err}
	}
	answer, err := s.askPortal(ctx, cfg, &portalRefreshReq{Key: keyText, InstallationID: installationID})
	if err != nil {
		return &portalAnswer{result: base.LicenseRefreshFailed, err: err}
	}
	return answer
}

// askPortal posts the key and the installation's ID to the portal's refresh, and
// reads its answer. A portal that cannot be reached, or is down or busy, is
// ErrLicensePortalUnavailable; one that refuses, ErrLicensePortalRefused. Only
// development may ask over plain http: the answer is not signed as a whole.
func (s *service) askPortal(ctx context.Context, cfg *config.Config, req *portalRefreshReq) (*portalAnswer, error) {
	refreshURL := cfg.LicensePortalRefreshURL()
	if u, err := url.Parse(refreshURL); err != nil || (u.Scheme != "https" && !cfg.IsDevEnv()) {
		return nil, hperrors.Wrap(hperrors.ErrLicensePortalUnavailable).
			WithExtraDetail("the portal's URL %s is not https, which only development allows", cfg.License.PortalURL)
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshURL, bytes.NewReader(body))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", portalUserAgent)

	res, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrLicensePortalUnavailable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(res.Body, maxPortalAnswerLen+1))
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrLicensePortalUnavailable).WithExtraDetail("%s", err.Error())
	}
	if len(data) > maxPortalAnswerLen {
		return nil, unreadableAnswer("longer than a key")
	}

	if res.StatusCode != http.StatusOK {
		return portalRefusal(res.StatusCode, data)
	}
	return readPortalAnswer(data)
}

// readPortalAnswer reads a 200: newer with a key, moved, or current.
func readPortalAnswer(data []byte) (*portalAnswer, error) {
	var resp portalRefreshResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, unreadableAnswer(err.Error())
	}
	if resp.Data == nil {
		return nil, unreadableAnswer("no data")
	}
	answer := &portalAnswer{result: resp.Data.Result}
	switch resp.Data.Result {
	case base.LicenseRefreshNewer:
		answer.newKey = license.NormalizeKeyText(resp.Data.Key)
		if answer.newKey == "" {
			return nil, unreadableAnswer("newer, with no key")
		}
	case base.LicenseRefreshMoved:
		if resp.Data.MovedAt != nil {
			answer.movedAt = resp.Data.MovedAt.UTC()
		}
	case base.LicenseRefreshCurrent:
	case base.LicenseRefreshUnknown, base.LicenseRefreshNone, base.LicenseRefreshDisabled,
		base.LicenseRefreshFailed:
		// The portal says unknown with a 404; the others are HivePaaS's own.
		fallthrough
	default:
		return nil, unreadableAnswer("a result of " + string(resp.Data.Result))
	}
	return answer, nil
}

// portalRefusal reads an answer other than 200. The portal's refresh answers 404,
// with its error and code, only for a key it did not issue: that is an answer,
// unknown. A 404 that is not the portal's error is a URL with no portal behind
// it. A portal busy or down is unavailable, and one that refuses the request -
// for another installation's key, say - refused.
func portalRefusal(status int, data []byte) (*portalAnswer, error) {
	var refusal portalError
	_ = json.Unmarshal(data, &refusal)
	if status == http.StatusNotFound && refusal.Code != "" {
		return &portalAnswer{result: base.LicenseRefreshUnknown}, nil
	}
	detail := http.StatusText(status)
	if refusal.Code != "" {
		detail = refusal.Code
	}
	if refusal.Detail != "" {
		detail += ": " + truncate(refusal.Detail, maxPortalDetailLen)
	}
	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		return nil, hperrors.Wrap(hperrors.ErrLicensePortalUnavailable).
			WithExtraDetail("the portal answered %d %s", status, detail)
	}
	return nil, hperrors.Wrap(hperrors.ErrLicensePortalRefused).
		WithExtraDetail("the portal answered %d %s", status, detail)
}

// unreadableAnswer is a 200 that is not an answer: a portal at that URL that is
// not one, or a broken one, which asking again later may find mended.
func unreadableAnswer(why string) error {
	return hperrors.Wrap(hperrors.ErrLicensePortalUnavailable).
		WithExtraDetail("the portal's answer could not be read: %s", why)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}
