package sessionuc

import (
	"context"
	"slices"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/openidConnect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) InitOAuthProvider(
	ctx context.Context,
	req *sessiondto.InitOAuthProviderReq,
) error {
	setting, oauth, err := uc.loadSSOSetting(ctx, req.Provider)
	if err != nil {
		return err
	}
	clientSecret, err := oauth.ClientSecret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	callbackURL := config.Current().SsoCallbackURL(req.Provider)

	var provider goth.Provider
	switch base.OAuthKind(setting.Kind) {
	case base.OAuthKindGithub, base.OAuthKindGithubApp:
		// goth asks GitHub for the verified primary email of somebody with no public one only when the scopes ask
		// for it: without, they cannot sign in. A GitHub App takes no scopes, and needs the Email addresses
		// permission instead, but goth decides by the scopes all the same.
		provider = github.New(oauth.ClientID, clientSecret, callbackURL, withEmailScope(oauth.Scopes)...)

	case base.OAuthKindGitlab:
		if oauth.AuthURL == "" {
			provider = gitlab.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)
		} else { // custom gitlab
			provider = gitlab.NewCustomisedURL(oauth.ClientID, clientSecret, callbackURL,
				oauth.AuthURL, oauth.TokenURL, oauth.ProfileURL, oauth.Scopes...)
		}

	case base.OAuthKindGitea:
		if oauth.AuthURL == "" {
			provider = gitea.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)
		} else { // a Gitea of one's own
			provider = gitea.NewCustomisedURL(oauth.ClientID, clientSecret, callbackURL,
				oauth.AuthURL, oauth.TokenURL, oauth.ProfileURL, oauth.Scopes...)
		}

	case base.OAuthKindGoogle:
		provider = google.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)

	case base.OAuthKindMicrosoftOnline:
		provider = microsoftonline.New(oauth.ClientID, clientSecret, callbackURL, oauth.Scopes...)

	case base.OAuthKindOpenIDConnect:
		provider, err = openidConnect.New(oauth.ClientID, clientSecret, callbackURL,
			oauth.AutoDiscoveryURL, oauth.Scopes...)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	if provider == nil {
		return hperrors.NewArgumentInvalid("OAuth kind").WithMsgLog("setting %s: OAuth kind %q", setting.ID,
			setting.Kind)
	}
	provider.SetName(req.Provider)
	goth.UseProviders(provider)

	return nil
}

// loadSSOSetting is the setting a sign-in with a provider uses, by its ID: an active OAuth one, or a GitHub App's
// with sign-in turned on. Nothing else is a provider, whatever ID the browser was given.
func (uc *UC) loadSSOSetting(ctx context.Context, id string) (*entity.Setting, *entity.OAuth, error) {
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, entity.NewObjectScopeGlobal(), "", id, true)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if setting.Type != base.SettingTypeOAuth && setting.Type != base.SettingTypeGithubApp {
		return nil, nil, hperrors.NewNotFound("OAuth").WithMsgLog("setting %s is of type %s", id, setting.Type)
	}
	oauth, err := setting.AsOAuth()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if oauth == nil { // a GitHub App that does not sign people in
		return nil, nil, hperrors.NewNotFound("OAuth").WithMsgLog("setting %s does not sign in", id)
	}
	return setting, oauth, nil
}

// githubScopeEmail lets goth read a GitHub account's emails, for the verified primary one.
const githubScopeEmail = "user:email"

// withEmailScope is scopes with githubScopeEmail added, unless it is there already.
func withEmailScope(scopes []string) []string {
	if slices.Contains(scopes, githubScopeEmail) {
		return scopes
	}
	return append(slices.Clone(scopes), githubScopeEmail)
}
