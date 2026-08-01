package envvarhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestCalcContentChanges(t *testing.T) {
	ev1 := []*entity.EnvVar{
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
		{Key: "SHARED_SECRET", Value: "xyz", IsBuild: false, IsShared: true},
		{Key: "BUILD_ARG", Value: "v1.0", IsBuild: true, IsShared: false},
	}

	// Case 1: Identical list (different order) -> no changes
	ev2 := []*entity.EnvVar{
		{Key: "BUILD_ARG", Value: "v1.0", IsBuild: true, IsShared: false},
		{Key: "SHARED_SECRET", Value: "xyz", IsBuild: false, IsShared: true},
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
	}

	runtimeChange, sharedChange, buildChange := CalcContentChanges(ev1, ev2)
	assert.False(t, runtimeChange)
	assert.False(t, sharedChange)
	assert.False(t, buildChange)

	// Case 2: Change runtime var value
	evRuntimeChanged := []*entity.EnvVar{
		{Key: "PORT", Value: "9090", IsBuild: false, IsShared: false},
		{Key: "SHARED_SECRET", Value: "xyz", IsBuild: false, IsShared: true},
		{Key: "BUILD_ARG", Value: "v1.0", IsBuild: true, IsShared: false},
	}
	runtimeChange, sharedChange, buildChange = CalcContentChanges(ev1, evRuntimeChanged)
	assert.True(t, runtimeChange)
	assert.False(t, sharedChange)
	assert.False(t, buildChange)

	// Case 3: Change shared var value
	evSharedChanged := []*entity.EnvVar{
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
		{Key: "SHARED_SECRET", Value: "new_value", IsBuild: false, IsShared: true},
		{Key: "BUILD_ARG", Value: "v1.0", IsBuild: true, IsShared: false},
	}
	runtimeChange, sharedChange, buildChange = CalcContentChanges(ev1, evSharedChanged)
	assert.True(t, runtimeChange) // Shared is also runtime (!IsBuild)
	assert.True(t, sharedChange)
	assert.False(t, buildChange)

	// Case 4: Change build var value
	evBuildChanged := []*entity.EnvVar{
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
		{Key: "SHARED_SECRET", Value: "xyz", IsBuild: false, IsShared: true},
		{Key: "BUILD_ARG", Value: "v2.0", IsBuild: true, IsShared: false},
	}
	runtimeChange, sharedChange, buildChange = CalcContentChanges(ev1, evBuildChanged)
	assert.False(t, runtimeChange)
	assert.False(t, sharedChange)
	assert.True(t, buildChange)
}
