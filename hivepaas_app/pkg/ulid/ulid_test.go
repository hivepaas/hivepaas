package ulid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewULID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ulid, err := NewULID()
		assert.Nil(t, err)
		assert.Equal(t, 16, len(ulid))
	})
}

func Test_NewStringULID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ulid, err := NewStringULID()
		assert.Nil(t, err)
		assert.Equal(t, 26, len(ulid))
	})

	t.Run("success with checking order", func(t *testing.T) {
		ulid1, err := NewStringULID()
		assert.Nil(t, err)
		assert.Equal(t, 26, len(ulid1))
		ulid2, err := NewStringULID()
		assert.Nil(t, err)
		assert.Equal(t, 26, len(ulid2))

		// ulid1 must less than ulid2
		assert.True(t, ulid1 < ulid2)
	})
}
