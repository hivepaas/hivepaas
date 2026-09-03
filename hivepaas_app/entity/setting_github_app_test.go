package entity

import "testing"

func TestGithubAppSettingsURL(t *testing.T) {
	tests := []struct {
		name string
		app  *GithubApp
		want string
	}{
		{
			name: "org-owned app uses the org settings page",
			app:  &GithubApp{Slug: "my-app", OwnerLogin: "acme", OwnerType: GithubAppOwnerTypeOrg},
			want: "https://github.com/organizations/acme/settings/apps/my-app",
		},
		{
			name: "user-owned app uses the personal settings page",
			app:  &GithubApp{Slug: "my-app", OwnerLogin: "alice", OwnerType: GithubAppOwnerTypeUser},
			want: "https://github.com/settings/apps/my-app",
		},
		{
			// Organization also means "restrict SSO to this org" and "default repo
			// owner", so it can be set on an app owned by a personal account.
			name: "user-owned app with an SSO organization still uses the personal page",
			app: &GithubApp{
				Slug: "my-app", OwnerLogin: "alice", OwnerType: GithubAppOwnerTypeUser,
				Organization: "acme",
			},
			want: "https://github.com/settings/apps/my-app",
		},
		{
			// The owner reported by GitHub wins over whatever was typed in the form.
			name: "owner from the API wins over the Organization field",
			app: &GithubApp{
				Slug: "my-app", OwnerLogin: "real-org", OwnerType: GithubAppOwnerTypeOrg,
				Organization: "typo-org",
			},
			want: "https://github.com/organizations/real-org/settings/apps/my-app",
		},
		{
			name: "legacy setting without an owner falls back to Organization",
			app:  &GithubApp{Slug: "my-app", Organization: "acme"},
			want: "https://github.com/organizations/acme/settings/apps/my-app",
		},
		{
			name: "unknown slug yields no URL rather than a guessed one",
			app:  &GithubApp{OwnerLogin: "acme", OwnerType: GithubAppOwnerTypeOrg},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.SettingsURL(); got != tt.want {
				t.Errorf("SettingsURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGithubAppOtherURLs(t *testing.T) {
	app := &GithubApp{Slug: "my-app", OwnerLogin: "acme", OwnerType: GithubAppOwnerTypeOrg}

	if got, want := app.InstallationsURL(),
		"https://github.com/organizations/acme/settings/apps/my-app/installations"; got != want {
		t.Errorf("InstallationsURL() = %q, want %q", got, want)
	}
	// The public page does not depend on the owner.
	if got, want := app.PublicURL(), "https://github.com/apps/my-app"; got != want {
		t.Errorf("PublicURL() = %q, want %q", got, want)
	}

	empty := &GithubApp{OwnerLogin: "acme"}
	if empty.InstallationsURL() != "" || empty.PublicURL() != "" {
		t.Error("URLs must be empty when the slug is unknown")
	}
}
