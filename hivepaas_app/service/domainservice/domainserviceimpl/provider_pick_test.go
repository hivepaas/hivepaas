package domainserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

var providerPickT0 = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

func providerSetting(id, owner string, isDefault bool, createdAfter time.Duration) *entity.Setting {
	return &entity.Setting{ID: id, ObjectID: owner, Default: isDefault, CreatedAt: providerPickT0.Add(createdAfter)}
}

func projectScope() *entity.ObjectScope {
	return &entity.ObjectScope{ScopeType: base.ObjectScopeProject, ProjectID: "p1"}
}

func TestPreferredSettingOfNoneIsNone(t *testing.T) {
	assert.Nil(t, preferredSetting(nil, projectScope()))
}

// The project's own provider is the one it chose, over any the installation has,
// even one the installation marked default.
func TestPreferredSettingTakesTheNearestScope(t *testing.T) {
	global := providerSetting("global", "", true, 0)
	project := providerSetting("project", "p1", false, time.Hour)

	assert.Equal(t, "project", preferredSetting([]*entity.Setting{global, project}, projectScope()).ID)
}

func TestPreferredSettingTakesTheDefaultAmongTheSameScope(t *testing.T) {
	older := providerSetting("older", "", false, 0)
	chosen := providerSetting("chosen", "", true, time.Hour)

	assert.Equal(t, "chosen", preferredSetting([]*entity.Setting{older, chosen}, projectScope()).ID)
	assert.Equal(t, "chosen", preferredSetting([]*entity.Setting{chosen, older}, projectScope()).ID)
}

// Without a default the answer is still the same one every time, whatever
// order the database returns them in: the oldest.
func TestPreferredSettingTakesTheOldestOtherwise(t *testing.T) {
	first := providerSetting("first", "", false, 0)
	second := providerSetting("second", "", false, time.Hour)

	assert.Equal(t, "first", preferredSetting([]*entity.Setting{second, first}, projectScope()).ID)
	assert.Equal(t, "first", preferredSetting([]*entity.Setting{first, second}, projectScope()).ID)
}

func TestScopeDistanceRisesToTheInstallation(t *testing.T) {
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "a1", ProjectEnvID: "e1", ProjectID: "p1"}

	assert.Equal(t, 0, scopeDistance(providerSetting("", "a1", false, 0), scope))
	assert.Less(t, scopeDistance(providerSetting("", "e1", false, 0), scope),
		scopeDistance(providerSetting("", "p1", false, 0), scope))
	assert.Less(t, scopeDistance(providerSetting("", "p1", false, 0), scope),
		scopeDistance(providerSetting("", "", false, 0), scope))
}
