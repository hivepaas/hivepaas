package schedjobuc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestFunctionCallsLiveInApps(t *testing.T) {
	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeApp, base.SchedJobTypeFunctionInvoke))
	for _, scope := range []base.ObjectScopeType{base.ObjectScopeGlobal, base.ObjectScopeProject,
		base.ObjectScopeProjectEnv} {
		err := checkJobTypeInScope(scope, base.SchedJobTypeFunctionInvoke)
		assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "%s: got %v", scope, err)
	}
}

func kindSetting(t *testing.T, category base.AppCategory) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{Type: base.SettingTypeAppKind}
	assert.NoError(t, setting.SetData(&entity.AppKindSettings{Category: category}))
	return setting
}

// A function is called in its own app, and only a function is.
func TestAFunctionCallIsItsFunctionsOwn(t *testing.T) {
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app-1"}
	job := &entity.SchedJob{JobType: base.SchedJobTypeFunctionInvoke, App: entity.ObjectID{ID: "app-1"},
		FunctionInvoke: &entity.SchedJobFunctionInvoke{Method: "GET", Path: "/"}}
	function := kindSetting(t, base.AppCategoryFunction)
	assert.NoError(t, checkFunctionInvokeApp(scope, job, function))

	other := *job
	other.App.ID = "app-2"
	err := checkFunctionInvokeApp(scope, &other, function)
	assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "another app: got %v", err)

	for _, kind := range []*entity.Setting{nil, kindSetting(t, base.AppCategoryWebapp)} {
		err = checkFunctionInvokeApp(scope, job, kind)
		assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "not a function: got %v", err)
	}

	assert.NoError(t, checkFunctionInvokeApp(scope, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand},
		nil), "another type is not checked")
}
