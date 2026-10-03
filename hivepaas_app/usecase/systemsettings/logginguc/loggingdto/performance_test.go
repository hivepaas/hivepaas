package loggingdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A node is listed once, with a capacity of the ones there are; none is
// auto.
func TestUpdateLoggingPerformanceReq(t *testing.T) {
	req := &UpdateLoggingPerformanceReq{Enabled: true, Nodes: []*PerformanceNodeReq{
		{ID: "n1"}, {ID: "n2", Capacity: "large"}, {ID: "n3", Capacity: "auto"},
	}}
	assert.Empty(t, req.Validate())
	assert.Equal(t, &entity.LoggingPerformance{Enabled: true, Nodes: []*entity.LoggingPerformanceNode{
		{ID: "n1", Capacity: "auto"}, {ID: "n2", Capacity: "large"}, {ID: "n3", Capacity: "auto"},
	}}, req.ToEntity())

	for name, nodes := range map[string][]*PerformanceNodeReq{
		"twice":         {{ID: "n1"}, {ID: "n1"}},
		"no id":         {{ID: ""}},
		"no capacity":   {{ID: "n1", Capacity: "huge"}},
		"a scale level": {{ID: "n1", Capacity: "-2"}},
	} {
		req := &UpdateLoggingPerformanceReq{Nodes: nodes}
		assert.NotEmpty(t, req.Validate(), name)
	}

	assert.Len(t, PerformanceCapacities(), 3)
	assert.Equal(t, "small", PerformanceCapacities()[0].Capacity)
	assert.Equal(t, 100, PerformanceCapacities()[0].MemoryMiB)
}
