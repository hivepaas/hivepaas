package sessionuc

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
)

const (
	// oauthEmailCheckTimeout bounds the one extra request a provider can need to tell whether an email is verified.
	oauthEmailCheckTimeout = 15 * time.Second
	// oauthEmailListMaxBytes bounds the answer to it: a list of a person's addresses, never large.
	oauthEmailListMaxBytes = 1 << 20

	// msGuestMarker is in the userPrincipalName Microsoft gives a guest of another tenant: the address before it is
	// the guest's own, which the tenant that invited them never verified.
	msGuestMarker = "#ext#"
)

// oauthHTTPClient makes the request verifiedOAuthEmail needs for Gitea; a test points it at its own server.
var oauthHTTPClient = &http.Client{Timeout: oauthEmailCheckTimeout}

// verifiedOAuthEmail is the email a sign-in with a provider may be matched to a user by: one the provider vouches the
// person proved they hold. An email the provider only reports - one anybody can type into a profile - would let whoever
// set it sign in as the HivePaaS user who has it.
//
// What vouches for it differs by provider:
//   - GitHub hands over the public email, which it lets be only a verified one, or the primary email when it is
//     verified (goth refuses it otherwise).
//   - Google says so in verified_email.
//   - GitLab gives the primary email, and confirmed_at once it is confirmed.
//   - Gitea's profile does not say, so its list of the person's emails is asked.
//   - Microsoft's mail is any text a tenant's administrator sets, and the app signs in people from every tenant. The
//     userPrincipalName is in a domain the tenant proved it owns, so it is taken, and mail only in that same domain.
//   - An OpenID Connect provider says so in email_verified, unless its setting trusts every email it gives: an
//     identity provider whose users and emails the administrator manages.
func (uc *UC) verifiedOAuthEmail(
	ctx context.Context,
	kind base.OAuthKind,
	oauth *entity.OAuth,
	user *goth.User,
) (string, error) {
	email := strutil.NormalizeEmail(user.Email)
	notVerified := func(reason string) (string, error) {
		return "", hperrors.Wrap(hperrors.ErrOAuthEmailNotVerified).
			WithMsgLog("sign-in with %s (%s): %s", kind, user.Provider, reason)
	}

	switch kind {
	case base.OAuthKindGithub, base.OAuthKindGithubApp:
		if email == "" {
			return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
		}
		return email, nil

	case base.OAuthKindGoogle:
		if email == "" {
			return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
		}
		if !claimTrue(user.RawData["verified_email"]) && !claimTrue(user.RawData["email_verified"]) {
			return notVerified("verified_email is not true")
		}
		return email, nil

	case base.OAuthKindGitlab:
		if email == "" {
			return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
		}
		if confirmedAt, _ := user.RawData["confirmed_at"].(string); strings.TrimSpace(confirmedAt) == "" {
			return notVerified("the account's email is not confirmed")
		}
		return email, nil

	case base.OAuthKindGitea:
		if email == "" {
			return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
		}
		verified, err := giteaEmailVerified(ctx, giteaEmailsURL(oauth), user.AccessToken, email)
		if err != nil {
			return "", hperrors.Wrap(hperrors.ErrOAuthEmailNotVerified).WithCause(err).
				WithMsgLog("sign-in with %s (%s): the emails could not be listed", kind, user.Provider)
		}
		if !verified {
			return notVerified("the email is not a verified one of the account")
		}
		return email, nil

	case base.OAuthKindMicrosoftOnline:
		return microsoftEmail(user.RawData, notVerified)

	case base.OAuthKindOpenIDConnect:
		if email == "" {
			return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
		}
		if oauth.TrustEmail {
			return email, nil
		}
		if !claimTrue(user.RawData["email_verified"]) {
			return notVerified("email_verified is not true")
		}
		return email, nil
	}
	return notVerified("an unknown kind of provider")
}

// microsoftEmail is the userPrincipalName of a Microsoft account, or its mail when that is in the same domain.
func microsoftEmail(raw map[string]any, notVerified func(string) (string, error)) (string, error) {
	upn, _ := raw["userPrincipalName"].(string)
	upn = strutil.NormalizeEmail(upn)
	if upn == "" {
		return "", hperrors.Wrap(hperrors.ErrOAuthUserEmailNotReturned)
	}
	if strings.Contains(upn, msGuestMarker) {
		return notVerified("a guest of the tenant")
	}
	upnDomain, ok := emailDomain(upn)
	if !ok {
		return notVerified("the userPrincipalName is no email address")
	}
	mail, _ := raw["mail"].(string)
	mail = strutil.NormalizeEmail(mail)
	if mailDomain, ok := emailDomain(mail); ok && mailDomain == upnDomain {
		return mail, nil
	}
	return upn, nil
}

// giteaEmailsURL is where the Gitea the setting signs in with lists a person's emails: beside its profile endpoint.
func giteaEmailsURL(oauth *entity.OAuth) string {
	profileURL := gitea.ProfileURL
	if oauth.AuthURL != "" && oauth.ProfileURL != "" {
		profileURL = oauth.ProfileURL
	}
	return strings.TrimRight(profileURL, "/") + "/emails"
}

// giteaEmailVerified asks Gitea, with the token the sign-in just got, whether email is a verified one of the person.
func giteaEmailVerified(ctx context.Context, emailsURL, accessToken, email string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, emailsURL, nil)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, hperrors.NewInternal().WithMsgLog("gitea answered %d listing the emails", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
	}
	if err = json.NewDecoder(http.MaxBytesReader(nil, resp.Body, oauthEmailListMaxBytes)).Decode(&emails); err != nil {
		return false, hperrors.Wrap(err)
	}
	for _, item := range emails {
		if item.Verified && strutil.NormalizeEmail(item.Email) == email {
			return true, nil
		}
	}
	return false, nil
}

// claimTrue reports whether a claim says true: a JSON boolean, or the string some providers send.
func claimTrue(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true")
	}
	return false
}

// emailDomain is the part after the @ of an address with exactly one, and something on each side.
func emailDomain(email string) (string, bool) {
	local, domain, found := strings.Cut(email, "@")
	if !found || local == "" || domain == "" || strings.Contains(domain, "@") {
		return "", false
	}
	return domain, true
}
