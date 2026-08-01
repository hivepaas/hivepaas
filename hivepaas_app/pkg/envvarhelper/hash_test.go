package envvarhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestCalcHashes(t *testing.T) {
	vars := []*entity.EnvVar{
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
		{Key: "APP_ENV", Value: "production", IsBuild: false, IsShared: true},
		{Key: "NODE_ENV", Value: "production", IsBuild: true, IsShared: false},
		{Key: "DATABASE_URL", Value: "postgres://localhost:5432/db", IsBuild: false, IsShared: true},
	}

	runtimeHash1, sharedHash1, buildHash1 := CalcHashes(vars)

	assert.NotEmpty(t, runtimeHash1)
	assert.NotEmpty(t, sharedHash1)
	assert.NotEmpty(t, buildHash1)

	// Order independence test: shuffle input slice
	shuffledVars := []*entity.EnvVar{
		{Key: "DATABASE_URL", Value: "postgres://localhost:5432/db", IsBuild: false, IsShared: true},
		{Key: "NODE_ENV", Value: "production", IsBuild: true, IsShared: false},
		{Key: "PORT", Value: "8080", IsBuild: false, IsShared: false},
		{Key: "APP_ENV", Value: "production", IsBuild: false, IsShared: true},
	}

	runtimeHash2, sharedHash2, buildHash2 := CalcHashes(shuffledVars)

	// Hashes must match regardless of input slice ordering
	assert.Equal(t, runtimeHash1, runtimeHash2)
	assert.Equal(t, sharedHash1, sharedHash2)
	assert.Equal(t, buildHash1, buildHash2)

	// Value change test
	changedVars := []*entity.EnvVar{
		{Key: "PORT", Value: "9090", IsBuild: false, IsShared: false}, // changed value
		{Key: "APP_ENV", Value: "production", IsBuild: false, IsShared: true},
		{Key: "NODE_ENV", Value: "production", IsBuild: true, IsShared: false},
		{Key: "DATABASE_URL", Value: "postgres://localhost:5432/db", IsBuild: false, IsShared: true},
	}

	runtimeHash3, sharedHash3, buildHash3 := CalcHashes(changedVars)
	assert.NotEqual(t, runtimeHash1, runtimeHash3)
	assert.Equal(t, sharedHash1, sharedHash3)
	assert.Equal(t, buildHash1, buildHash3)
}

func TestCalcHashes_Empty(t *testing.T) {
	runtimeHash, sharedHash, buildHash := CalcHashes(nil)
	assert.Empty(t, runtimeHash)
	assert.Empty(t, sharedHash)
	assert.Empty(t, buildHash)
}
