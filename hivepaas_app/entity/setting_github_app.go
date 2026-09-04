package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	CurrentGithubAppVersion = 1

	// Owner types as reported by the GitHub API (the `owner.type` field).
	GithubAppOwnerTypeOrg  = "Organization"
	GithubAppOwnerTypeUser = "User"

	githubBaseURL = "https://github.com"
)

var _ = registerSettingParser(base.SettingTypeGithubApp, &githubAppParser{})

type githubAppParser struct {
}

func (s *githubAppParser) New() SettingData {
	return &GithubApp{}
}

type GithubApp struct {
	ClientID     string         `json:"clientId"`
	ClientSecret EncryptedField `json:"clientSecret"`
	Organization string         `json:"org"`
	// Slug is the URL-friendly app identifier GitHub generates from the name at
	// creation time. It must be read from the API, never derived from the name:
	// the normalization is lossy and GitHub appends a counter on collisions
	// ("my app" can become "my-app-2"), and it does not follow later renames.
	Slug string `json:"slug"`
	// OwnerLogin and OwnerType identify who actually owns the app on GitHub, as
	// reported by the API. They must not be derived from Organization: that field
	// is free text on the manual create/update path and is overloaded to also mean
	// the SSO restriction org and the default repo owner, so it can be non-empty
	// on a user-owned app (and wrong on an org-owned one).
	OwnerLogin     string         `json:"ownerLogin"`
	OwnerType      string         `json:"ownerType"`
	WebhookURL     string         `json:"webhookURL"`
	WebhookSecret  EncryptedField `json:"webhookSecret"`
	AppID          int64          `json:"appId"`
	InstallationID int64          `json:"installationId"`
	PrivateKey     EncryptedField `json:"privateKey"`
	SSOEnabled     bool           `json:"ssoEnabled"`
}

// IsOrgOwned reports whether the app belongs to an organization rather than a
// user account. The two have different settings URLs on github.com.
func (s *GithubApp) IsOrgOwned() bool {
	if s.OwnerType != "" {
		return s.OwnerType == GithubAppOwnerTypeOrg
	}
	// Settings created before the owner was recorded. The manifest flow only ever
	// creates an app under an org when Organization is set, so this is right for
	// apps it created - but not necessarily for ones registered by hand.
	return s.Organization != ""
}

// ownerLoginForURL returns the account the settings URL must be built under.
func (s *GithubApp) ownerLoginForURL() string {
	if !s.IsOrgOwned() {
		return "" // personal settings page, no owner segment
	}
	if s.OwnerLogin != "" {
		return s.OwnerLogin
	}
	return s.Organization // legacy settings, before the owner was recorded
}

// GithubAppOwnerSettingsBaseURL returns the github.com settings base URL of the
// account that owns (or will own) an app: the org page when org is set, the
// personal settings page otherwise.
func GithubAppOwnerSettingsBaseURL(org string) string {
	if org != "" {
		return githubBaseURL + "/organizations/" + org + "/settings/apps"
	}
	return githubBaseURL + "/settings/apps"
}

// SettingsURL returns the github.com settings page of this app, e.g.
// https://github.com/organizations/acme/settings/apps/my-hivepaas-app
//
// It returns an empty string when the slug is unknown (settings created before
// the slug was persisted); refresh it from the API in that case.
func (s *GithubApp) SettingsURL() string {
	if s.Slug == "" {
		return ""
	}
	return GithubAppOwnerSettingsBaseURL(s.ownerLoginForURL()) + "/" + s.Slug
}

// InstallationsURL returns the page listing where this app is installed.
func (s *GithubApp) InstallationsURL() string {
	settingsURL := s.SettingsURL()
	if settingsURL == "" {
		return ""
	}
	return settingsURL + "/installations"
}

// PermissionsURL returns the page listing permissions and subscribed events of the app.
func (s *GithubApp) PermissionsURL() string {
	settingsURL := s.SettingsURL()
	if settingsURL == "" {
		return ""
	}
	return settingsURL + "/permissions"
}

// PublicURL returns the public page of the app, which is where a user installs
// it. Unlike SettingsURL it does not depend on the owner type.
func (s *GithubApp) PublicURL() string {
	if s.Slug == "" {
		return ""
	}
	return githubBaseURL + "/apps/" + s.Slug
}

// CarryOverFrom copies the fields an update request does not carry.
//
// Updating a setting replaces its data wholesale, so every field absent from the
// request is silently lost unless it is copied here. Losing WebhookSecret breaks
// the signature check on every delivery, and losing Slug/OwnerLogin/OwnerType
// breaks the github.com links - all without any error surfacing.
func (s *GithubApp) CarryOverFrom(current *GithubApp) {
	if current == nil {
		return
	}
	s.WebhookURL = current.WebhookURL
	s.WebhookSecret = current.WebhookSecret
	s.Slug = current.Slug
	s.OwnerLogin = current.OwnerLogin
	s.OwnerType = current.OwnerType
}

func (s *GithubApp) GetType() base.SettingType {
	return base.SettingTypeGithubApp
}

func (s *GithubApp) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *GithubApp) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *GithubApp) Decrypt() error {
	_, err := s.ClientSecret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = s.WebhookSecret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = s.PrivateKey.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *GithubApp) ConvertAsOAuth() *OAuth {
	if !s.SSOEnabled {
		return nil
	}
	return &OAuth{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		Organization: s.Organization,
	}
}

func (s *GithubApp) ConvertAsRepoWebhook() *RepoWebhook {
	return &RepoWebhook{
		Kind:   base.WebhookKindGithub,
		Secret: s.WebhookSecret,
	}
}

func (s *Setting) AsGithubApp() (*GithubApp, error) {
	return parseSettingAs[*GithubApp](s)
}

func (s *Setting) MustAsGithubApp() *GithubApp {
	return gofn.Must(s.AsGithubApp())
}
