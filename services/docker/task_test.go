package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// An app never deployed has no service. Asked about it, Docker would read the
// empty filter as a prefix of every service; it is not asked at all.
func TestTasksOfNoServiceAreNone(t *testing.T) {
	m := &manager{} // no client: reaching Docker would panic
	for _, id := range []string{"", "  "} {
		resp, err := m.ServiceTaskList(context.Background(), id, nil)
		assert.NoError(t, err)
		if assert.NotNil(t, resp) {
			assert.Empty(t, resp.Items)
		}
	}
}
