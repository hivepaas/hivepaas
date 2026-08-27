package gitapi

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

func TestAccessTokenConn(
	ctx context.Context,
	kind base.AccessTokenKind,
	token string,
	baseURL string,
) (err error) {
	switch kind { //nolint:exhaustive
	case base.AccessTokenKindGithub:
		err = testGithubTokenConn(ctx, token, baseURL)
	case base.AccessTokenKindGitlab:
		err = testGitlabTokenConn(ctx, token, baseURL)
	case base.AccessTokenKindGitea:
		err = testGiteaTokenConn(ctx, token, baseURL)
	case base.AccessTokenKindBitbucket, base.AccessTokenKindGogs:
		fallthrough
	default:
		err = hperrors.Wrap(hperrors.ErrTokenTypeUnsupported).WithParam("Type", kind)
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func testGithubTokenConn(
	ctx context.Context,
	token string,
	_ string,
) error {
	client, err := github.NewFromToken(token)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, _, err = client.ListUserRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func testGitlabTokenConn(
	ctx context.Context,
	token string,
	baseURL string,
) error {
	client, err := gitlab.NewFromToken(token, baseURL)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, _, err = client.ListAllProjects(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func testGiteaTokenConn(
	ctx context.Context,
	token string,
	baseURL string,
) error {
	client, err := gitea.NewFromToken(token, baseURL)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, _, err = client.ListAllRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
