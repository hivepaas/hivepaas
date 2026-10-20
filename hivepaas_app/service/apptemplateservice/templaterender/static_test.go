package templaterender

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A version's tree has its override applied and nothing substituted; the
// template stays as it was. A deprecated version has one too: an app that runs
// it is updated, never created.
func TestVersionTree(t *testing.T) {
	tmpl := pgTemplate(t)

	tree, err := VersionTree(tmpl, "18")

	assert.NoError(t, err)
	mounts := tree.(map[string]any)["deployment"].(map[string]any)["storage"].(map[string]any)["mounts"]
	assert.Contains(t, mounts, "/var/lib/postgresql")
	assert.NotContains(t, mounts, "/var/lib/postgresql/data")
	assert.Equal(t, "${{ params.dataVolume }}", mounts.(map[string]any)["/var/lib/postgresql"].(map[string]any)["source"])
	templateMounts := tmpl.App["deployment"].(map[string]any)["storage"].(map[string]any)["mounts"]
	assert.Contains(t, templateMounts, "/var/lib/postgresql/data", "the template is not modified")

	_, err = VersionTree(tmpl, "16")
	assert.NoError(t, err, "deprecated")
	_, err = VersionTree(tmpl, "9")
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVersionNotFound)
}

// A part of a template that may hold no placeholder: escapes are undone, and a
// placeholder is refused.
func TestStatic(t *testing.T) {
	out, err := Static(map[string]any{"content": "echo ${HOME} $${{ not.one }}", "n": 3})

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"content": "echo ${HOME} ${{ not.one }}", "n": 3}, out)

	_, err = Static([]any{"a", "${{ params.password }}"})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}
