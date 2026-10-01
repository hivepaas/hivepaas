package webhookuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestPRCommentsRefusal(t *testing.T) {
	owner := &entity.App{ID: "app1", Name: "shop-api", ProjectEnvID: "prj1:prod"}
	const dashboard = "https://hivepaas.example.com"
	const previewsOff = "Preview deployments are disabled"

	for name, tc := range map[string]struct {
		settings *entity.AppFeaturePreviewSettings
		want     string
	}{
		"no preview settings":   {nil, previewsOff},
		"previews off":          {&entity.AppFeaturePreviewSettings{AllowPRComments: true}, previewsOff},
		"comments not allowed":  {&entity.AppFeaturePreviewSettings{Enabled: true}, "Allow PR Comments"},
		"previews and comments": {&entity.AppFeaturePreviewSettings{Enabled: true, AllowPRComments: true}, ""},
	} {
		got := prCommentsRefusal(owner, tc.settings, dashboard)
		if tc.want == "" {
			assert.Empty(t, got, name)
			continue
		}
		assert.Contains(t, got, tc.want, name)
		assert.Contains(t, got, "shop-api", name)
	}

	// The refusal of comments links to where they are allowed.
	got := prCommentsRefusal(owner, &entity.AppFeaturePreviewSettings{Enabled: true}, dashboard)
	assert.Contains(t, got, "(https://hivepaas.example.com/projects/prj1/prod/apps/app1/feature-settings/)")
	assert.Contains(t, got, "/hivepaas cancel", "cancel is refused too")
}

func TestFeatureSettingsURL(t *testing.T) {
	app := &entity.App{ID: "app1", ProjectEnvID: "prj1:prod"}
	assert.Equal(t, "https://hp.example.com/projects/prj1/prod/apps/app1/feature-settings/",
		featureSettingsURL("https://hp.example.com/", app))
	assert.Empty(t, featureSettingsURL("", app), "no dashboard address, no link")
	assert.Empty(t, featureSettingsURL("https://hp.example.com", &entity.App{ID: "app1"}))

	// Without a link, the refusal still says where.
	got := buildPRCommentsDisabledComment("shop-api", "")
	assert.Contains(t, got, "**Feature Settings**")
	assert.NotContains(t, got, "](")
}
