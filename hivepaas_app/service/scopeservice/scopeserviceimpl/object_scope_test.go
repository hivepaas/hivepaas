package scopeserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// A scope type that is not in the enum reaches LoadObjectScope whenever it comes from stored data
// instead of a validated request. It has to come back as an error, not a nil-pointer panic.
func TestLoadObjectScope_UnknownScopeType(t *testing.T) {
	t.Parallel()

	s := &service{} // the unknown-type path must fail before it touches any dependency

	scope, err := s.LoadObjectScope(context.Background(), nil, base.ObjectScopeType("bogus"), "id-1", true)

	assert.Error(t, err)
	assert.Nil(t, scope)
}
