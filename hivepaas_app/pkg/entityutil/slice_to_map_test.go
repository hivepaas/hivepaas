package entityutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceToMap(t *testing.T) {
	entities := []*mockEntity{
		{ID: "id1", Name: "Name1"},
		{ID: "id2", Name: "Name2"},
	}

	t.Run("SliceToIDMap", func(t *testing.T) {
		res := SliceToIDMap(entities)
		assert.Len(t, res, 2)
		assert.Equal(t, entities[0], res["id1"])
		assert.Equal(t, entities[1], res["id2"])
	})

	t.Run("SliceToNameMap - Case Sensitive", func(t *testing.T) {
		res := SliceToNameMap(entities, true)
		assert.Len(t, res, 2)
		assert.Equal(t, entities[0], res["Name1"])
		assert.Equal(t, entities[1], res["Name2"])
		assert.Nil(t, res["name1"])
	})

	t.Run("SliceToNameMap - Case Insensitive", func(t *testing.T) {
		res := SliceToNameMap(entities, false)
		assert.Len(t, res, 2)
		assert.Equal(t, entities[0], res["name1"])
		assert.Equal(t, entities[1], res["name2"])
	})
}
