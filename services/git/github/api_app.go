package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// GetApp returns the metadata of the authenticated GitHub App (GET /app).
//
// This is the only reliable way to learn the app's slug: it is generated from
// the name at creation time through a lossy normalization, GitHub appends a
// counter on collisions, and it does not follow later renames. Deriving it from
// the app name gives a wrong URL as soon as either happens.
//
// The call is authenticated with the app JWT, so it needs no installation.
func (c *Client) GetApp(ctx context.Context) (*gogithub.App, error) {
	if !c.IsAppClient() {
		return nil, hperrors.Wrap(ErrGithubAppClientRequired)
	}

	app, _, err := c.appClient.Apps.Get(ctx, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return app, nil
}
