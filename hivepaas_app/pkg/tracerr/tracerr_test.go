package tracerr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Wrap(t *testing.T) {
	var nilErr error
	assert.Nil(t, Wrap(nilErr))
	assert.Nil(t, Wrap(nil))

	e1 := errors.New("my error")
	e2 := Wrap(e1)
	e3 := Wrap(e2)
	assert.ErrorIs(t, e2, e1)
	assert.ErrorIs(t, e3, e2)
}
