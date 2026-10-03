package webhookuc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Who may run commands: a writer, by GitHub's word, by owning the repository,
// or by the provider's answer; with none of these, a private repository's
// commenters only.
func TestPRCommentAuthorVerdict(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name     string
		author   prCommentAuthor
		canWrite *bool
		want     bool
	}{
		{"github owner", prCommentAuthor{Association: "OWNER"}, nil, true},
		{"github org member", prCommentAuthor{Association: "MEMBER"}, nil, true},
		{"github collaborator", prCommentAuthor{Association: "COLLABORATOR"}, nil, true},
		{"github contributor", prCommentAuthor{Association: "CONTRIBUTOR"}, nil, false},
		{"github first-time contributor", prCommentAuthor{Association: "FIRST_TIME_CONTRIBUTOR"}, nil, false},
		{"github stranger", prCommentAuthor{Association: "NONE"}, nil, false},
		{"github stranger, private repo", prCommentAuthor{Association: "NONE", RepoPrivate: true}, nil, false},
		{"repo owner", prCommentAuthor{IsRepoOwner: true}, nil, true},
		{"provider says writer", prCommentAuthor{}, &yes, true},
		{"provider says not, private repo", prCommentAuthor{RepoPrivate: true}, &no, false},
		{"no answer, public repo", prCommentAuthor{}, nil, false},
		{"no answer, private repo", prCommentAuthor{RepoPrivate: true}, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, prCommentAuthorVerdict(&c.author, c.canWrite))
		})
	}
}

func TestPRCommentAuthorNeedsAsking(t *testing.T) {
	assert.True(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Login: "u"}, base.WebhookKindGitea))
	assert.True(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Login: "u"}, base.WebhookKindGitlab))
	assert.False(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Login: "u", IsRepoOwner: true}, base.WebhookKindGitea),
		"the owner may: nothing to ask")
	assert.False(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Association: "NONE"}, base.WebhookKindGithub))
	assert.False(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Login: "u"}, base.WebhookKindBitbucket),
		"no API to ask")
	assert.False(t, prCommentAuthorNeedsAsking(&prCommentAuthor{Login: "u"}, base.WebhookKindGogs))
}

// Each provider's comment webhook gives the comment's author.
func TestParsePRCommentAuthor(t *testing.T) {
	// GitLab's own spelling of the field naming what a note is on.
	const gitlabNoteOn = "note" + "able_type"
	cases := []struct {
		name    string
		kind    base.WebhookKind
		headers map[string]string
		body    string
		want    prCommentAuthor
	}{
		{
			name:    "github",
			kind:    base.WebhookKindGithub,
			headers: map[string]string{"X-GitHub-Event": "issue_comment"},
			body: `{"action":"created","issue":{"number":7,"pull_request":{"url":"x"}},
				"comment":{"body":"/hivepaas deploy","user":{"login":"stranger"},"author_association":"NONE"},
				"repository":{"html_url":"https://github.com/acme/web"}}`,
			want: prCommentAuthor{Login: "stranger", Association: "NONE"},
		},
		{
			name:    "gitea, the owner",
			kind:    base.WebhookKindGitea,
			headers: map[string]string{"X-Gitea-Event": "issue_comment"},
			body: `{"action":"created","is_pull":true,"issue":{"number":7},
				"comment":{"body":"/hivepaas deploy","user":{"login":"Acme"}},
				"repository":{"html_url":"https://gitea.example.com/acme/web","private":true,"owner":{"login":"acme"}}}`,
			want: prCommentAuthor{Login: "Acme", IsRepoOwner: true, RepoPrivate: true},
		},
		{
			name:    "gitlab",
			kind:    base.WebhookKindGitlab,
			headers: map[string]string{"X-Gitlab-Event": "Note Hook"},
			body: `{"object_kind":"note","user":{"id":42,"username":"dev"},
				"object_attributes":{"note":"/hivepaas deploy","` + gitlabNoteOn + `":"MergeRequest"},
				"merge_request":{"iid":7},"repository":{"git_http_url":"https://gitlab.com/acme/web.git"}}`,
			want: prCommentAuthor{Login: "dev", ID: 42},
		},
		{
			name:    "bitbucket, someone else on a public repository",
			kind:    base.WebhookKindBitbucket,
			headers: map[string]string{"X-Event-Key": "pullrequest:comment_created"},
			body: `{"actor":{"nickname":"stranger","uuid":"{u2}"},
				"repository":{"is_private":false,"owner":{"uuid":"{u1}"},
					"links":{"html":{"href":"https://bitbucket.org/acme/web"}}},
				"pullrequest":{"id":7,"source":{"branch":{"name":"fix"}}},
				"comment":{"content":{"raw":"/hivepaas deploy"}}}`,
			want: prCommentAuthor{Login: "stranger"},
		},
		{
			name:    "gogs, someone else on a private repository",
			kind:    base.WebhookKindGogs,
			headers: map[string]string{"X-Gogs-Event": "issue_comment"},
			body: `{"action":"created","issue":{"number":7,"pull_request":{}},
				"comment":{"body":"/hivepaas deploy"},"sender":{"username":"dev"},
				"repository":{"html_url":"https://gogs.example.com/acme/web","private":true,
					"owner":{"username":"acme"}}}`,
			want: prCommentAuthor{Login: "dev", RepoPrivate: true},
		},
	}
	uc := &UC{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(c.body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			data, err := uc.parseRepoWebhook(req, c.kind, "")
			if assert.NoError(t, err) && assert.NotNil(t, data.PRComment) {
				assert.Equal(t, c.want, data.PRComment.Author)
			}
		})
	}
}
