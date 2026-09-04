package webhookuc

import (
	"net/http"
	"slices"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func requestWithHeaders(headers map[string]string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/", nil) //nolint:noctx
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	return req
}

func TestDetectWebhookKind(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    base.WebhookKind
	}{
		{
			name:    "github",
			headers: map[string]string{headerGithubEvent: "push"},
			want:    base.WebhookKindGithub,
		},
		{
			name:    "gitlab",
			headers: map[string]string{headerGitlabEvent: "Push Hook"},
			want:    base.WebhookKindGitlab,
		},
		{
			name:    "bitbucket",
			headers: map[string]string{headerBitbucketEvent: "repo:push"},
			want:    base.WebhookKindBitbucket,
		},
		{
			// Gitea sends GitHub's header too, for compatibility. Reading it as
			// GitHub would check the delivery against the wrong signature scheme.
			name: "gitea wins over the github header it also sends",
			headers: map[string]string{
				headerGiteaEvent:  "push",
				headerGithubEvent: "push",
			},
			want: base.WebhookKindGitea,
		},
		{
			name: "gogs wins over the github header it also sends",
			headers: map[string]string{
				headerGogsEvent:   "push",
				headerGithubEvent: "push",
			},
			want: base.WebhookKindGogs,
		},
		{
			name:    "no known header yields no kind",
			headers: map[string]string{"X-Something-Else": "push"},
			want:    "",
		},
		{
			name:    "an empty header value does not count",
			headers: map[string]string{headerGithubEvent: ""},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectWebhookKind(requestWithHeaders(tt.headers)); got != tt.want {
				t.Errorf("detectWebhookKind() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectWebhookKindNilRequest(t *testing.T) {
	if got := detectWebhookKind(nil); got != "" {
		t.Errorf("a nil request should yield no kind, got %q", got)
	}
}

// A webhook with a kind configured must use exactly that parser: no detection, no
// fallback, so the sender's headers cannot steer it onto another verification
// scheme - Bitbucket's in particular compares the secret from a header instead of
// signing the body.
func TestWebhookCandidatesConfiguredKindIsNeverRetried(t *testing.T) {
	// Headers claiming to be every other provider at once.
	req := requestWithHeaders(map[string]string{
		headerGiteaEvent:     "push",
		headerGogsEvent:      "push",
		headerGitlabEvent:    "Push Hook",
		headerBitbucketEvent: "repo:push",
		headerGithubEvent:    "push",
	})

	for _, kind := range base.AllWebhookKinds {
		t.Run(string(kind), func(t *testing.T) {
			got := webhookCandidates(req, kind)
			if len(got) != 1 || got[0] != kind {
				t.Errorf("expected only %q, got %v", kind, got)
			}
		})
	}
}

func TestWebhookCandidatesFallbackOnlyForUnsetKind(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    []base.WebhookKind
	}{
		{
			// The ambiguous case: Gitea sends its own header and GitHub's, so both
			// are worth trying, Gitea first.
			name: "gitea falls back to github",
			headers: map[string]string{
				headerGiteaEvent:  "push",
				headerGithubEvent: "push",
			},
			want: []base.WebhookKind{base.WebhookKindGitea, base.WebhookKindGithub},
		},
		{
			name: "gogs falls back to github",
			headers: map[string]string{
				headerGogsEvent:   "push",
				headerGithubEvent: "push",
			},
			want: []base.WebhookKind{base.WebhookKindGogs, base.WebhookKindGithub},
		},
		{
			// No overlap: nothing to fall back to.
			name:    "github alone has no fallback",
			headers: map[string]string{headerGithubEvent: "push"},
			want:    []base.WebhookKind{base.WebhookKindGithub},
		},
		{
			name:    "gitlab alone has no fallback",
			headers: map[string]string{headerGitlabEvent: "Push Hook"},
			want:    []base.WebhookKind{base.WebhookKindGitlab},
		},
		{
			// Bitbucket authenticates by comparing the secret from a header, so it
			// must never be tried unless it actually announced itself.
			name:    "bitbucket is never a fallback for another provider",
			headers: map[string]string{headerGithubEvent: "push"},
			want:    []base.WebhookKind{base.WebhookKindGithub},
		},
		{
			name:    "an unknown sender yields no candidate",
			headers: map[string]string{"X-Something-Else": "push"},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := webhookCandidates(requestWithHeaders(tt.headers), "")
			if !slices.Equal(got, tt.want) {
				t.Errorf("webhookCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebhookCandidatesPutsTheDetectedKindFirst(t *testing.T) {
	req := requestWithHeaders(map[string]string{
		headerGiteaEvent:  "push",
		headerGithubEvent: "push",
	})
	got := webhookCandidates(req, "")
	if len(got) == 0 || got[0] != base.WebhookKindGitea {
		t.Errorf("the detected kind must be tried first, got %v", got)
	}
}
