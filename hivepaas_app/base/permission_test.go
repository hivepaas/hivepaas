package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AnyOf is how a check asks for one of several actions - upload asks for write,
// and a session limited to reading must be refused it. It used to answer the
// opposite: yes when one of the actions was missing, no when all were there.
func TestAllowsAnyIsTrueWhenOneActionIsAllowed(t *testing.T) {
	readOnly := &AccessActions{Read: true}
	full := &AccessActions{Read: true, Exec: true, Write: true, Del: true}
	none := &AccessActions{}

	assert.False(t, readOnly.AllowsAny([]ActionType{ActionTypeWrite}))
	assert.True(t, readOnly.AllowsAny([]ActionType{ActionTypeWrite, ActionTypeRead}))
	assert.True(t, full.AllowsAny([]ActionType{ActionTypeWrite}))
	assert.False(t, none.AllowsAny([]ActionType{ActionTypeRead, ActionTypeWrite}))
	assert.False(t, (*AccessActions)(nil).AllowsAny([]ActionType{ActionTypeRead}))
}

func TestAllowsAllNeedsEveryAction(t *testing.T) {
	readOnly := &AccessActions{Read: true}

	assert.True(t, readOnly.AllowsAll([]ActionType{ActionTypeRead}))
	assert.False(t, readOnly.AllowsAll([]ActionType{ActionTypeRead, ActionTypeWrite}))
}
