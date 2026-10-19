package appmanageddto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A point's name is held to its engine's limit, which the request is checked
// against once the app, and so its engine, is known: 63 for Postgres.
func TestAPointNameIsHeldToTheEnginesLimit(t *testing.T) {
	for name, tc := range map[string]struct {
		name  string
		valid bool
	}{
		"a name":        {"before-v42", true},
		"63 bytes":      {strings.Repeat("a", 63), true},
		"64 bytes":      {strings.Repeat("a", 64), false},
		"none":          {"", false},
		"spaces around": {"  ", false},
	} {
		req := &CreateRestorePointReq{Name: tc.name}
		assert.NoError(t, req.ModifyRequest())
		errs := req.ValidateName(63)
		assert.Equal(t, tc.valid, len(errs) == 0, name)
	}
}

func TestARestoreTargetPointIsHeldToTheEnginesLimit(t *testing.T) {
	point := func(name string) *RestoreManagedReq {
		return &RestoreManagedReq{Target: ManagedRestoreTargetReq{Kind: entity.ManagedRestoreToPoint, Point: name}}
	}

	assert.Empty(t, point(strings.Repeat("a", 63)).ValidatePoint(63))
	assert.Len(t, point(strings.Repeat("a", 64)).ValidatePoint(63), 1)
	assert.Empty(t, (&RestoreManagedReq{Target: ManagedRestoreTargetReq{Kind: entity.ManagedRestoreToLatest}}).
		ValidatePoint(63), "no point")
}
