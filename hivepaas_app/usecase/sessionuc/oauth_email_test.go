package sessionuc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// Each provider vouches for an email its own way; an email it only reports is refused.
func TestVerifiedOAuthEmailByProvider(t *testing.T) {
	cases := map[string]struct {
		kind  base.OAuthKind
		oauth *entity.OAuth
		user  goth.User
		want  string
		err   error
	}{
		"github": {kind: base.OAuthKindGithub, user: goth.User{Email: "Jane@Example.com"}, want: "jane@example.com"},
		"github app": {kind: base.OAuthKindGithubApp, user: goth.User{Email: "jane@example.com"},
			want: "jane@example.com"},
		"github, no email": {kind: base.OAuthKindGithub, err: hperrors.ErrOAuthUserEmailNotReturned},

		"google, verified": {kind: base.OAuthKindGoogle, user: goth.User{Email: "jane@example.com",
			RawData: map[string]any{"verified_email": true}}, want: "jane@example.com"},
		"google, email_verified": {kind: base.OAuthKindGoogle, user: goth.User{Email: "jane@example.com",
			RawData: map[string]any{"email_verified": "true"}}, want: "jane@example.com"},
		"google, not verified": {kind: base.OAuthKindGoogle, user: goth.User{Email: "jane@example.com",
			RawData: map[string]any{"verified_email": false}}, err: hperrors.ErrOAuthEmailNotVerified},
		"google, not said": {kind: base.OAuthKindGoogle, user: goth.User{Email: "jane@example.com"},
			err: hperrors.ErrOAuthEmailNotVerified},

		"gitlab, confirmed": {kind: base.OAuthKindGitlab, user: goth.User{Email: "jane@example.com",
			RawData: map[string]any{"confirmed_at": "2024-05-23T09:05:22Z"}}, want: "jane@example.com"},
		"gitlab, unconfirmed": {kind: base.OAuthKindGitlab, user: goth.User{Email: "jane@example.com",
			RawData: map[string]any{"confirmed_at": nil}}, err: hperrors.ErrOAuthEmailNotVerified},

		"microsoft, mail in the UPN's domain": {kind: base.OAuthKindMicrosoftOnline, user: goth.User{
			RawData: map[string]any{"userPrincipalName": "j.doe@contoso.com", "mail": "Jane@Contoso.com"}},
			want: "jane@contoso.com"},
		"microsoft, mail elsewhere": {kind: base.OAuthKindMicrosoftOnline, user: goth.User{
			RawData: map[string]any{"userPrincipalName": "mallory@evil.onmicrosoft.com", "mail": "jane@contoso.com"}},
			want: "mallory@evil.onmicrosoft.com"},
		"microsoft, no mail": {kind: base.OAuthKindMicrosoftOnline, user: goth.User{
			RawData: map[string]any{"userPrincipalName": "jane@contoso.com"}}, want: "jane@contoso.com"},
		"microsoft, a guest": {kind: base.OAuthKindMicrosoftOnline, user: goth.User{
			RawData: map[string]any{"userPrincipalName": "jane_example.com#EXT#@evil.onmicrosoft.com",
				"mail": "jane@example.com"}}, err: hperrors.ErrOAuthEmailNotVerified},
		"microsoft, no UPN": {kind: base.OAuthKindMicrosoftOnline, user: goth.User{
			RawData: map[string]any{"mail": "jane@contoso.com"}}, err: hperrors.ErrOAuthUserEmailNotReturned},

		"oidc, verified": {kind: base.OAuthKindOpenIDConnect, oauth: &entity.OAuth{}, user: goth.User{
			Email: "jane@example.com", RawData: map[string]any{"email_verified": true}}, want: "jane@example.com"},
		"oidc, not verified": {kind: base.OAuthKindOpenIDConnect, oauth: &entity.OAuth{}, user: goth.User{
			Email: "jane@example.com", RawData: map[string]any{"email_verified": false}},
			err: hperrors.ErrOAuthEmailNotVerified},
		"oidc, not said": {kind: base.OAuthKindOpenIDConnect, oauth: &entity.OAuth{},
			user: goth.User{Email: "jane@example.com"}, err: hperrors.ErrOAuthEmailNotVerified},
		"oidc, trusted": {kind: base.OAuthKindOpenIDConnect, oauth: &entity.OAuth{TrustEmail: true},
			user: goth.User{Email: "jane@example.com"}, want: "jane@example.com"},

		"an unknown kind": {kind: "myspace", user: goth.User{Email: "jane@example.com"},
			err: hperrors.ErrOAuthEmailNotVerified},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			oauth := tc.oauth
			if oauth == nil {
				oauth = &entity.OAuth{}
			}
			email, err := (&UC{}).verifiedOAuthEmail(context.Background(), tc.kind, oauth, &tc.user)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				assert.Empty(t, email)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, email)
		})
	}
}

// Gitea's profile does not say whether its email is verified: its list of the person's emails is asked, with the
// sign-in's token, at the Gitea the setting names.
func TestVerifiedOAuthEmailAsksGitea(t *testing.T) {
	var emails []map[string]any
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/user/emails", r.URL.Path)
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(emails)
	}))
	defer server.Close()
	oauth := &entity.OAuth{AuthURL: server.URL + "/login/oauth/authorize",
		TokenURL: server.URL + "/login/oauth/access_token", ProfileURL: server.URL + "/api/v1/user"}
	user := &goth.User{Email: "jane@example.com", AccessToken: "tok"}
	check := func() (string, error) {
		return (&UC{}).verifiedOAuthEmail(context.Background(), base.OAuthKindGitea, oauth, user)
	}

	emails = []map[string]any{{"email": "Jane@Example.com", "verified": true, "primary": true}}
	email, err := check()
	require.NoError(t, err)
	assert.Equal(t, "jane@example.com", email)

	emails = []map[string]any{{"email": "jane@example.com", "verified": false, "primary": true},
		{"email": "other@example.com", "verified": true}}
	_, err = check()
	require.ErrorIs(t, err, hperrors.ErrOAuthEmailNotVerified)

	status, emails = http.StatusForbidden, nil
	_, err = check()
	require.ErrorIs(t, err, hperrors.ErrOAuthEmailNotVerified, "the list not given: nothing vouches")
}

func TestGiteaEmailsURL(t *testing.T) {
	assert.Equal(t, "https://gitea.com/api/v1/user/emails", giteaEmailsURL(&entity.OAuth{}))
	assert.Equal(t, "https://git.acme.example/api/v1/user/emails", giteaEmailsURL(&entity.OAuth{
		AuthURL: "https://git.acme.example/login/oauth/authorize", ProfileURL: "https://git.acme.example/api/v1/user/"}))
}

func TestWithEmailScope(t *testing.T) {
	scopes := []string{"read:user"}
	assert.Equal(t, []string{"read:user", "user:email"}, withEmailScope(scopes))
	assert.Equal(t, []string{"read:user"}, scopes, "the setting's own scopes stay as they are")
	assert.Equal(t, []string{"user:email"}, withEmailScope([]string{"user:email"}))
	assert.Equal(t, []string{"user:email"}, withEmailScope(nil))
}

type settingByID struct {
	repository.SettingRepo
	settings map[string]*entity.Setting
}

func (r *settingByID) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if setting, found := r.settings[id]; found {
		return setting, nil
	}
	return nil, hperrors.NewNotFound("Setting")
}

type userByEmail struct {
	repository.UserRepo
	users map[string]*entity.User
}

func (r *userByEmail) GetByEmail(_ context.Context, _ database.IDB, email string,
	_ ...bunex.SelectQueryOption) (*entity.User, error) {
	if user, found := r.users[email]; found {
		return user, nil
	}
	return nil, hperrors.NewNotFound("User")
}

func (r *userByEmail) Update(context.Context, database.IDB, *entity.User, ...bunex.UpdateQueryOption) error {
	return nil
}

func ssoSettings(t *testing.T) map[string]*entity.Setting {
	t.Helper()
	google := &entity.Setting{ID: "set-google", Type: base.SettingTypeOAuth, Kind: string(base.OAuthKindGoogle)}
	require.NoError(t, google.SetData(&entity.OAuth{ClientID: "cid"}))
	ghApp := &entity.Setting{ID: "set-ghapp", Type: base.SettingTypeGithubApp, Kind: string(base.OAuthKindGithubApp)}
	require.NoError(t, ghApp.SetData(&entity.GithubApp{ClientID: "cid", SSOEnabled: false}))
	email := &entity.Setting{ID: "set-email", Type: base.SettingTypeEmail, Kind: "smtp"}
	return map[string]*entity.Setting{google.ID: google, ghApp.ID: ghApp, email.ID: email}
}

func newOAuthUCTest(t *testing.T) (*UC, *spyAuditService) {
	t.Helper()
	uc, audit := newAuditUCTest()
	uc.settingRepo = &settingByID{settings: ssoSettings(t)}
	uc.userRepo = &userByEmail{users: map[string]*entity.User{
		"jane@example.com": {ID: "u-jane", Username: "jane", Email: "jane@example.com", Status: base.UserStatusActive},
	}}
	return uc, audit
}

// An email the provider does not vouch for signs nobody in, and the user who has it sees the attempt in their
// history; an email nobody has is refused as such.
func TestCreateOAuthSessionRefusals(t *testing.T) {
	t.Run("not verified", func(t *testing.T) {
		uc, audit := newOAuthUCTest(t)
		_, err := uc.CreateOAuthSession(context.Background(), &sessiondto.CreateOAuthSessionReq{User: &goth.User{
			Provider: "set-google", Email: "Jane@example.com", RawData: map[string]any{"verified_email": false}}})
		require.ErrorIs(t, err, hperrors.ErrOAuthEmailNotVerified)
		require.Len(t, audit.entries, 1)
		entry := audit.entries[0]
		assert.Equal(t, base.AuditLogResultDenied, entry.Result)
		assert.Equal(t, "u-jane", entry.ResID)
		assert.Equal(t, auditMethodOAuth, entry.Section)
		assert.Contains(t, entry.Detail, auditReasonEmailNotVerified)
	})

	t.Run("nobody has it", func(t *testing.T) {
		uc, audit := newOAuthUCTest(t)
		_, err := uc.CreateOAuthSession(context.Background(), &sessiondto.CreateOAuthSessionReq{User: &goth.User{
			Provider: "set-google", Email: "nobody@example.com", RawData: map[string]any{"verified_email": true}}})
		require.ErrorIs(t, err, hperrors.ErrOAuthUserNotFound)
		require.Len(t, audit.entries, 1)
		assert.Empty(t, audit.entries[0].ResID)
		assert.Contains(t, audit.entries[0].Detail, auditReasonUnknownUser)
	})
}

// Only an OAuth setting, or a GitHub App's that signs people in, is a provider: any other ID is not found, rather
// than read as one.
func TestLoadSSOSetting(t *testing.T) {
	uc, _ := newOAuthUCTest(t)
	_, oauth, err := uc.loadSSOSetting(context.Background(), "set-google")
	require.NoError(t, err)
	assert.Equal(t, "cid", oauth.ClientID)

	for _, id := range []string{"set-ghapp", "set-email", "set-missing"} {
		_, _, err = uc.loadSSOSetting(context.Background(), id)
		require.ErrorIs(t, err, hperrors.ErrNotFound, id)
	}
}
