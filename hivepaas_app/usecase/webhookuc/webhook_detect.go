package webhookuc

import (
	"bytes"
	"io"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// Event headers each provider identifies itself with. They match what the
// go-playground/webhooks parsers read, so a kind detected here is one that
// parser can actually handle.
const (
	headerGithubEvent    = "X-GitHub-Event"
	headerGitlabEvent    = "X-Gitlab-Event"
	headerGiteaEvent     = "X-Gitea-Event"
	headerGogsEvent      = "X-Gogs-Event"
	headerBitbucketEvent = "X-Event-Key"
)

// detectWebhookKind identifies the sender of a delivery from its event header.
//
// A webhook with no kind configured accepts any provider, which is what the
// default webhook created with a project does: the provider is only known once a
// delivery arrives.
//
// Order matters. Gitea and Gogs also send GitHub's X-GitHub-Event for
// compatibility, so their own headers must be tested first - otherwise their
// deliveries would be parsed, and signature-checked, as GitHub and always fail.
// GitHub is therefore the last one tried.
//
// Detection picks the parser only; each parser still verifies the delivery
// against the webhook's own secret, so this cannot be used to skip that check.
// webhookCandidates returns the parsers to try for a delivery, most likely first.
//
// A webhook with a kind configured is never guessed at and never retried: it gets
// exactly that parser, so an explicitly configured provider cannot be talked into
// a different verification scheme by the sender's headers.
//
// Only an unset kind - a webhook that accepts any provider - falls back, and only
// to providers that sent their own event header. That matters because Bitbucket
// authenticates by comparing X-Hook-UUID against the secret rather than signing
// the body; trying it unconditionally would make that weaker scheme reachable for
// every webhook.
func webhookCandidates(req *http.Request, kind base.WebhookKind) []base.WebhookKind {
	if kind != "" {
		return []base.WebhookKind{kind}
	}

	detected := detectWebhookKind(req)
	if detected == "" {
		return nil
	}

	candidates := []base.WebhookKind{detected}
	// Gitea, Gogs and GitHub share X-GitHub-Event, so detection between them is a
	// guess. Keep the others as a fallback rather than letting a wrong guess
	// reject a delivery outright.
	for _, other := range base.AllWebhookKinds {
		if other != detected && hasOwnEventHeader(req, other) {
			candidates = append(candidates, other)
		}
	}
	return candidates
}

// hasOwnEventHeader reports whether the request carries the event header of that
// provider specifically.
func hasOwnEventHeader(req *http.Request, kind base.WebhookKind) bool {
	header, ok := eventHeaderByKind[kind]
	if !ok {
		return false
	}
	return req.Header.Get(header) != ""
}

var eventHeaderByKind = map[base.WebhookKind]string{
	base.WebhookKindGitea:     headerGiteaEvent,
	base.WebhookKindGogs:      headerGogsEvent,
	base.WebhookKindGitlab:    headerGitlabEvent,
	base.WebhookKindBitbucket: headerBitbucketEvent,
	base.WebhookKindGithub:    headerGithubEvent,
}

func detectWebhookKind(req *http.Request) base.WebhookKind {
	if req == nil {
		return ""
	}
	switch {
	case req.Header.Get(headerGiteaEvent) != "":
		return base.WebhookKindGitea
	case req.Header.Get(headerGogsEvent) != "":
		return base.WebhookKindGogs
	case req.Header.Get(headerGitlabEvent) != "":
		return base.WebhookKindGitlab
	case req.Header.Get(headerBitbucketEvent) != "":
		return base.WebhookKindBitbucket
	case req.Header.Get(headerGithubEvent) != "":
		return base.WebhookKindGithub
	default:
		return ""
	}
}

// parseRepoWebhook verifies and parses a delivery with the parser of the
// webhook's provider.
//
// A webhook that accepts any provider may have to try more than one parser -
// Gitea and Gogs also send GitHub's event header, so identifying the sender from
// headers alone is a guess. Retrying keeps a wrong guess from rejecting a valid
// delivery. A webhook with a kind configured gets that one parser and nothing
// else, so no header can steer it onto a different verification scheme.
//
// The body has to be buffered because each attempt reads it to completion; the
// handler caps its size beforehand.
func (uc *UC) parseRepoWebhook(
	req *http.Request,
	kind base.WebhookKind,
	secret string,
) (*repoEventData, error) {
	candidates := webhookCandidates(req, kind)
	if len(candidates) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrWebhookTypeUnsupported).WithParam("Type", kind)
	}

	var body []byte
	if len(candidates) > 1 && req.Body != nil {
		buffered, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		_ = req.Body.Close()
		body = buffered
	}

	var lastErr error
	for idx, candidate := range candidates {
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		eventData := &repoEventData{}
		err := uc.parseRepoWebhookAs(req, candidate, secret, eventData)
		if err == nil {
			return eventData, nil
		}
		lastErr = err

		if idx == len(candidates)-1 {
			break
		}
		logging.Warnf("webhook: the %s parser rejected the delivery, trying the next provider: %v",
			candidate, err)
	}
	return nil, hperrors.Wrap(lastErr)
}

func (uc *UC) parseRepoWebhookAs(
	req *http.Request,
	kind base.WebhookKind,
	secret string,
	eventData *repoEventData,
) error {
	switch kind {
	case base.WebhookKindGithub:
		return uc.parseGithubWebhook(req, secret, eventData)
	case base.WebhookKindGitlab:
		return uc.parseGitlabWebhook(req, secret, eventData)
	case base.WebhookKindGitea:
		return uc.parseGiteaWebhook(req, secret, eventData)
	case base.WebhookKindBitbucket:
		return uc.parseBitbucketWebhook(req, secret, eventData)
	case base.WebhookKindGogs:
		return uc.parseGogsWebhook(req, secret, eventData)
	default:
		return hperrors.Wrap(hperrors.ErrWebhookTypeUnsupported).WithParam("Type", kind)
	}
}
