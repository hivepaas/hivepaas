package appdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func appOfKind(t *testing.T, category base.AppCategory) *entity.App {
	t.Helper()
	app := &entity.App{ID: "app-1", ProjectEnv: &entity.ProjectEnv{Name: "dev"}}
	if category != "" {
		kind := &entity.Setting{Type: base.SettingTypeAppKind, Status: base.SettingStatusActive}
		assert.NoError(t, kind.SetData(&entity.AppKindSettings{Category: category}))
		app.Settings = []*entity.Setting{kind}
	}
	return app
}

// An app says what it is - a function, a database - so that a list can show
// it; one that declares no kind says nothing.
func TestAnAppSaysWhatItIs(t *testing.T) {
	function, err := TransformApp(appOfKind(t, base.AppCategoryFunction), nil)
	assert.NoError(t, err)
	assert.Equal(t, base.AppCategoryFunction, function.Category)

	plain, err := TransformApp(appOfKind(t, ""), nil)
	assert.NoError(t, err)
	assert.Empty(t, plain.Category)
}

func TestAListIsFilteredByKnownCategories(t *testing.T) {
	req := &ListAppReq{ProjectID: "01J0000000000000000000PRJ1",
		Category: []base.AppCategory{base.AppCategoryFunction}}
	assert.Empty(t, req.Validate())

	req.Category = []base.AppCategory{"queue"}
	assert.NotEmpty(t, req.Validate())
}
