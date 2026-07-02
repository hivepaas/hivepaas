package entityutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockEntity struct {
	ID   string
	Name string
}

func (e *mockEntity) GetID() string {
	return e.ID
}

func (e *mockEntity) GetName() string {
	return e.Name
}

func TestExtractIDs(t *testing.T) {
	entities := []*mockEntity{
		{ID: "id1", Name: "name1"},
		{ID: "id2", Name: "name2"},
	}

	ids := ExtractIDs(entities)
	assert.Equal(t, []string{"id1", "id2"}, ids)
}
