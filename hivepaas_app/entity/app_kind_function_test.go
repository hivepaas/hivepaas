package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func kindSetting(t *testing.T, category base.AppCategory) *Setting {
	t.Helper()
	setting := &Setting{ID: "kind-1", Type: base.SettingTypeAppKind, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(&AppKindSettings{Category: category}))
	return setting
}

// An app is a function by its kind.
func TestAnAppIsAFunctionByItsKind(t *testing.T) {
	assert.True(t, IsFunctionKind(kindSetting(t, base.AppCategoryFunction)))
	assert.False(t, IsFunctionKind(kindSetting(t, base.AppCategoryWebapp)))
	assert.False(t, IsFunctionKind(nil))
	assert.False(t, IsFunctionKind(&Setting{Type: base.SettingTypeAppRouting}))
}
