package imagescanservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func ignoreAt(scope base.ObjectScopeType, inheritable bool, vulnID string) *entity.Setting {
	setting := &entity.Setting{Type: base.SettingTypeVulnerabilityIgnore, Status: base.SettingStatusActive,
		Scope: scope, Inheritable: inheritable}
	setting.MustSetData(&entity.VulnerabilityIgnore{VulnID: vulnID, Reason: "r"})
	return setting
}

// A part of HivePaaS is of no project: the global ignores apply to it, those
// the apps inherit and those they do not; no other scope's.
func TestComponentIgnores(t *testing.T) {
	ignores := ComponentIgnores([]*entity.Setting{
		ignoreAt(base.ObjectScopeGlobal, true, "CVE-1"),
		ignoreAt(base.ObjectScopeGlobal, false, "CVE-2"),
		ignoreAt(base.ObjectScopeProject, true, "CVE-3"),
		ignoreAt(base.ObjectScopeApp, false, "CVE-4"),
	})
	var ids []string
	for _, ignore := range ignores {
		ids = append(ids, ignore.Data.VulnID)
	}
	assert.Equal(t, []string{"CVE-1", "CVE-2"}, ids)

	// The apps inherit only the first.
	appIgnores := IgnoresOf(&entity.App{ID: "a"}, []*entity.Setting{ignoreAt(base.ObjectScopeGlobal, false, "CVE-2")})
	assert.Empty(t, appIgnores)
}

// A part's results are kept under an ID no app has; each part is found by one
// way.
func TestComponents(t *testing.T) {
	traefik, found := ComponentByKey("traefik")
	assert.True(t, found)
	assert.Equal(t, "hivepaas:traefik", traefik.ID())
	assert.True(t, IsComponentID(traefik.ID()))
	assert.False(t, IsComponentID("01J9ZZ"))
	_, found = ComponentByKey("nope")
	assert.False(t, found)

	keys := map[string]bool{}
	for _, c := range Components {
		ways := 0
		for _, way := range []string{c.Service, c.SystemAppKey, c.Container} {
			if way != "" {
				ways++
			}
		}
		assert.Equal(t, 1, ways, c.Key)
		assert.False(t, keys[c.Key], "one key once: %s", c.Key)
		keys[c.Key] = true
	}
}
