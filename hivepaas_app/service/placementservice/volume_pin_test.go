package placementservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolumePinConstraintByID(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

func TestVolumePinConstraintByLabel(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeLabel: "storage=fast"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.labels.storage==fast", constraint)
}

// A label with no value means the key has to be present, which swarm spells as
// the string "true".
func TestVolumePinConstraintBareLabel(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{{VolumeName: "d", NodeLabel: "ssd"}})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.labels.ssd==true", constraint)
}

// Volumes with no pin say every node reaches their data, so they constrain
// nothing.
func TestVolumePinConstraintIgnoresUnpinned(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "shared"},
		{VolumeName: "pgdata", NodeID: "node-1"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

func TestVolumePinConstraintAgreeingPins(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-1"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

// Two volumes on two different nodes cannot both be reached by one task. This is
// a real contradiction, so it is reported rather than handed to swarm as a set
// of constraints no node satisfies.
func TestVolumePinConstraintConflict(t *testing.T) {
	_, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-2"},
	})

	assert.NotNil(t, conflict)
	assert.Contains(t, conflict.Error(), "pgdata")
	assert.Contains(t, conflict.Error(), "uploads")
}

func TestVolumePinConstraintConflictAcrossIDAndLabel(t *testing.T) {
	_, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeLabel: "storage=fast"},
	})

	assert.NotNil(t, conflict)
}
