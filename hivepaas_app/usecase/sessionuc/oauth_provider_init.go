package sessionuc

import (
	"context"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/openidConnect"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) InitOAuthProvider(
	ctx context.Context,
	req *sessiondto.InitOAuthProviderReq,
) error {
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, base.NewObjectScopeGlobal(), "", req.Provider, true)
	if err != nil {
		return apperrors.New(err)
	}

	oauth := setting.MustAsOAuth()
	clientSecret, err := oauth.ClientSecret.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}
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

	case base.OAuthKindMicrosoftOnline:
		provider = microsoftonline.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)

	case base.OAuthKindOpenIDConnect:
		provider, err = openidConnect.New(oauth.ClientID, clientSecret, callbackURL,
			oauth.AutoDiscoveryURL, oauth.Scopes...)
		if err != nil {
			return apperrors.New(err)
		}
	}
	provider.SetName(req.Provider)
	goth.UseProviders(provider)

	return nil
}
