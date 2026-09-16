package hpappserviceimpl

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	testTemplatesCommit = "0123456789abcdef0123456789abcdef01234567"
	testIndexSHA256     = "674231f24d8fd88ad11f4b3c0f6e91c5bf96bc03de857920d6804d6c94534c3e"
)

func TestValidateTemplatesRef(t *testing.T) {
	valid := base.TemplatesRef{Repo: "hivepaas/app-templates", Commit: testTemplatesCommit, IndexSHA256: testIndexSHA256}
	assert.NoError(t, validateTemplatesRef(&valid))

	for name, mutate := range map[string]func(ref *base.TemplatesRef){
		"empty repo":          func(ref *base.TemplatesRef) { ref.Repo = "" },
		"repo without owner":  func(ref *base.TemplatesRef) { ref.Repo = "app-templates" },
		"repo as URL":         func(ref *base.TemplatesRef) { ref.Repo = "https://github.com/hivepaas/app-templates" },
		"repo with extra dir": func(ref *base.TemplatesRef) { ref.Repo = "hivepaas/app-templates/main" },
		"repo name ..":        func(ref *base.TemplatesRef) { ref.Repo = "hivepaas/.." },
		"short commit":        func(ref *base.TemplatesRef) { ref.Commit = testTemplatesCommit[:7] },
		"branch as commit":    func(ref *base.TemplatesRef) { ref.Commit = "main" },
		"uppercase commit":    func(ref *base.TemplatesRef) { ref.Commit = strings.ToUpper(testTemplatesCommit) },
		"empty index sha256":  func(ref *base.TemplatesRef) { ref.IndexSHA256 = "" },
		"index sha1":          func(ref *base.TemplatesRef) { ref.IndexSHA256 = testTemplatesCommit },
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			ref := valid
			mutate(&ref)
			assert.ErrorIs(t, validateTemplatesRef(&ref), hperrors.ErrReleaseInfoInvalid)
		})
	}
}

func TestDecodeReleaseInfo_Templates(t *testing.T) {
	t.Run("pin is decoded per channel", func(t *testing.T) {
		info, err := decodeReleaseInfo([]byte(`{
			"stable": {"appVersion": "v0.1.1", "templates": {"repo": "hivepaas/app-templates",
				"commit": "` + testTemplatesCommit + `", "indexSha256": "` + testIndexSHA256 + `"}},
			"beta": {"appVersion": "v0.1.1-beta3"}
		}`))
		assert.NoError(t, err)
		assert.Equal(t, &base.TemplatesRef{
			Repo: "hivepaas/app-templates", Commit: testTemplatesCommit, IndexSHA256: testIndexSHA256,
		}, info.Stable.Templates)
		assert.Nil(t, info.Beta.Templates, "a channel without a pin offers no templates")
	})

	t.Run("a malformed pin refuses the release info", func(t *testing.T) {
		info, err := decodeReleaseInfo([]byte(`{
			"stable": {"appVersion": "v0.1.1"},
			"beta": {"appVersion": "v0.1.1-beta3", "templates": {"repo": "hivepaas/app-templates",
				"commit": "main", "indexSha256": "` + testIndexSHA256 + `"}}
		}`))
		assert.Nil(t, info)
		assert.ErrorIs(t, err, hperrors.ErrReleaseInfoInvalid)
	})
}

// release.json is signed as it is, and a field the app refuses would stop every
// installation from seeing updates. Checking the committed file here catches that
// before it is signed rather than after it is published.
func TestRepoReleaseJSONDecodes(t *testing.T) {
	data, err := os.ReadFile("../../../../release.json")
	assert.NoError(t, err)
	_, err = decodeReleaseInfo(data)
	assert.NoError(t, err)
}
