package entity

import "testing"

// An update request carries only a subset of the fields and the setting data is
// replaced wholesale, so the rest must survive or they are silently lost.
func TestGithubAppCarryOverFrom(t *testing.T) {
	current := &GithubApp{
		WebhookURL:    "https://hive.example/api/webhooks/set_1",
		WebhookSecret: NewEncryptedField("s3cr3t"),
		Slug:          "my-app",
		OwnerLogin:    "acme",
		OwnerType:     GithubAppOwnerTypeOrg,
	}
	// What UpdateGithubAppReq.ToEntity() produces: no webhook, no owner, no slug.
	updated := &GithubApp{
		ClientID:     "new-client-id",
		Organization: "acme",
		AppID:        42,
	}

	updated.CarryOverFrom(current)

	if got := updated.WebhookSecret.String(); got != "s3cr3t" {
		t.Errorf("webhook secret was lost, got %q", got)
	}
	if updated.WebhookURL != current.WebhookURL {
		t.Errorf("webhook URL was lost, got %q", updated.WebhookURL)
	}
	if updated.Slug != "my-app" || updated.OwnerLogin != "acme" || updated.OwnerType != GithubAppOwnerTypeOrg {
		t.Errorf("owner/slug were lost: %+v", updated)
	}
	// Fields the request does carry must not be touched.
	if updated.ClientID != "new-client-id" || updated.AppID != 42 {
		t.Errorf("request fields were overwritten: %+v", updated)
	}
}

func TestGithubAppCarryOverFromNil(t *testing.T) {
	app := &GithubApp{ClientID: "keep-me"}
	app.CarryOverFrom(nil)
	if app.ClientID != "keep-me" {
		t.Error("a nil current setting must be a no-op")
	}
}
