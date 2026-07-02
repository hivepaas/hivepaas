package entityutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	entities := []*mockEntity{
		{ID: "id1", Name: "Name1"},
		{ID: "id2", Name: "Name2"},
	}

	t.Run("FindByID", func(t *testing.T) {
		assert.Equal(t, entities[0], FindByID(entities, "id1"))
		assert.Equal(t, entities[1], FindByID(entities, "id2"))
		assert.Nil(t, FindByID(entities, "id3"))
	})

	t.Run("FindByName", func(t *testing.T) {
		assert.Equal(t, entities[0], FindByName(entities, "Name1"))
		assert.Equal(t, entities[1], FindByName(entities, "Name2"))
		assert.Nil(t, FindByName(entities, "Name3"))
	})
}
