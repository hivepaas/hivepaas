package sessionuc

import (
	"context"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) InitOAuthProvider(
	ctx context.Context,
	req *sessiondto.InitOAuthProviderReq,
) error {
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, base.NewSettingScopeGlobal(), "", req.Provider, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	oauth := setting.MustAsOAuth()
	clientSecret := oauth.ClientSecret.MustGetPlain()
	callbackURL := config.Current.SsoCallbackURL(req.Provider)

	var provider goth.Provider
	switch base.OAuthKind(setting.Kind) {
	case base.OAuthKindGithub, base.OAuthKindGithubApp:
		provider = github.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)

	case base.OAuthKindGitlab:
		if oauth.AuthURL == "" {
			provider = gitlab.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)
		} else { // custom gitlab
			provider = gitlab.NewCustomisedURL(oauth.ClientID, clientSecret, callbackURL,
				oauth.AuthURL, oauth.TokenURL, oauth.ProfileURL, oauth.Scopes...)
		}

	case base.OAuthKindGitea:
		provider = gitea.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)

	case base.OAuthKindGoogle:
		provider = google.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)
	}
	provider.SetName(req.Provider)
	goth.UseProviders(provider)

	return nil
}
