package webhookuc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func githubPush(secret string) *http.Request {
	body := `{"ref":"refs/heads/main","after":"0123456789abcdef0123456789abcdef01234567",` +
		`"repository":{"html_url":"https://github.com/acme/shop"}}`
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body)) //nolint:noctx
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	return req
}

// A delivery the webhook's secret does not sign is the sender's fault, and
// answered so - not as the server's own failure.
func TestParseRepoWebhookRefusesWhatTheSecretDoesNotSign(t *testing.T) {
	uc := &UC{}
	for _, kind := range []base.WebhookKind{base.WebhookKindGithub, ""} {
		data, err := uc.parseRepoWebhook(githubPush("the-secret"), kind, "the-secret")
		if err != nil || data.Push == nil || data.Push.RepoRef != "refs/heads/main" {
			t.Fatalf("kind %q: a signed push = %+v, %v; want it read", kind, data, err)
		}
		for name, req := range map[string]*http.Request{
			"signed with another secret": githubPush("another-secret"),
			"not signed":                 githubPush(""),
		} {
			_, err := uc.parseRepoWebhook(req, kind, "the-secret")
			if !errors.Is(err, hperrors.ErrWebhookUnverified) {
				t.Errorf("kind %q, a push %s: err = %v; want ErrWebhookUnverified", kind, name, err)
			}
			if !errors.Is(err, hperrors.ErrUnauthorized) {
				t.Errorf("kind %q, a push %s: err = %v; want it unauthorized", kind, name, err)
			}
		}
	}
}
