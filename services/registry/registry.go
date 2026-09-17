package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	requestTimeout = 10 * time.Second
	// pageSize is what a registry is asked for at once. Docker Hub honors 1000,
	// and a repository the size of library/postgres - 1421 tags - is then two
	// requests instead of fifteen.
	pageSize    = 1000
	maxPages    = 10
	maxBodySize = 4 << 20
)

// challengeParamPattern reads key="value" out of a WWW-Authenticate header.
var challengeParamPattern = regexp.MustCompile(`([a-z]+)="([^"]*)"`)

type Client struct {
	http *http.Client
	// endpoint is the base URL for a registry host. It is a field so a test can
	// point the client at an httptest server instead of the real internet.
	endpoint func(host string) string
}

func New() *Client {
	return &Client{
		// The transport reads HTTP_PROXY, HTTPS_PROXY and NO_PROXY from the
		// environment, which is where config.Proxy puts them.
		http:     &http.Client{Timeout: requestTimeout},
		endpoint: func(host string) string { return "https://" + host },
	}
}

type ListTagsResult struct {
	Tags []string
	// Truncated says the registry had more tags than maxTags allowed.
	Truncated bool
}

// ListTags reads the tags of one repository, following the token challenge the
// registry answers with and the Link header it pages by.
//
// maxTags bounds the work: `postgres` alone publishes thousands of tags, and the
// caller only shows a handful of them.
func (c *Client) ListTags(ctx context.Context, ref Reference, maxTags int) (*ListTagsResult, error) {
	result := &ListTagsResult{Tags: make([]string, 0, pageSize)}
	url := fmt.Sprintf("%s/v2/%s/tags/list?n=%d", c.endpoint(ref.Host), ref.Name, pageSize)
	token := ""

	for page := 0; page < maxPages && url != ""; page++ {
		body, next, newToken, err := c.readPage(ctx, url, token, ref)
		if err != nil {
			return nil, err
		}
		token = newToken

		decoded := &struct {
			Tags []string `json:"tags"`
		}{}
		if err = json.Unmarshal(body, decoded); err != nil {
			return nil, hperrors.Wrap(hperrors.ErrRegistryUnavailable).
				WithExtraDetail("%s: %s", ref, err.Error())
		}
		for _, tag := range decoded.Tags {
			if len(result.Tags) >= maxTags {
				result.Truncated = true
				return result, nil
			}
			result.Tags = append(result.Tags, tag)
		}
		url = next
	}
	result.Truncated = result.Truncated || url != ""
	return result, nil
}

// readPage fetches one page, answering a 401 challenge once. The token it returns
// is reused for the next page, so a repository costs one token, not one per page.
func (c *Client) readPage(
	ctx context.Context,
	url, token string,
	ref Reference,
) (body []byte, next, newToken string, err error) {
	res, err := c.get(ctx, url, token)
	if err != nil {
		return nil, "", "", err
	}

	if res.StatusCode == http.StatusUnauthorized && token == "" {
		challenge := res.Header.Get("WWW-Authenticate")
		_ = res.Body.Close()
		if token, err = c.fetchToken(ctx, challenge, ref); err != nil {
			return nil, "", "", err
		}
		if res, err = c.get(ctx, url, token); err != nil {
			return nil, "", "", err
		}
	}
	defer func() { _ = res.Body.Close() }()

	if err = statusError(res, ref); err != nil {
		return nil, "", "", err
	}
	body, err = io.ReadAll(io.LimitReader(res.Body, maxBodySize))
	if err != nil {
		return nil, "", "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).
			WithExtraDetail("%s: %s", ref, err.Error())
	}
	return body, nextPageURL(res, c.endpoint(ref.Host)), token, nil
}

func (c *Client) get(ctx context.Context, url, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s", err.Error())
	}
	return res, nil
}

// fetchToken answers a Bearer challenge. Docker Hub issues an anonymous token for
// a public repository, which is why no credentials appear here.
func (c *Client) fetchToken(ctx context.Context, challenge string, ref Reference) (string, error) {
	params := map[string]string{}
	for _, match := range challengeParamPattern.FindAllStringSubmatch(challenge, -1) {
		params[match[1]] = match[2]
	}
	realm := params["realm"]
	if realm == "" {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).
			WithExtraDetail("%s: the registry asked for authentication without naming a realm", ref)
	}

	url := realm + "?scope=" + params["scope"]
	if service := params["service"]; service != "" {
		url += "&service=" + service
	}
	res, err := c.get(ctx, url, "")
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	if err = statusError(res, ref); err != nil {
		return "", err
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxBodySize))
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, err.Error())
	}
	decoded := &struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}{}
	if err = json.Unmarshal(body, decoded); err != nil {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, err.Error())
	}
	if decoded.Token == "" {
		return decoded.AccessToken, nil
	}
	return decoded.Token, nil
}

func statusError(res *http.Response, ref Reference) error {
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusTooManyRequests:
		return hperrors.Wrap(hperrors.ErrRegistryRateLimited).WithExtraDetail("%s", ref)
	case http.StatusNotFound:
		return hperrors.NewNotFound("Image repository").WithMsgLog("repository %v not found", ref)
	default:
		return hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, res.Status)
	}
}

// nextPageURL reads the Link header registries page with. The URL in it is a path,
// so it is joined back onto the host that answered.
func nextPageURL(res *http.Response, base string) string {
	link := res.Header.Get("Link")
	if link == "" || !strings.Contains(link, `rel="next"`) {
		return ""
	}
	start := strings.Index(link, "<")
	end := strings.Index(link, ">")
	if start < 0 || end <= start {
		return ""
	}
	target := link[start+1 : end]
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return base + target
}
