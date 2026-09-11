package placementserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

func applyData(pins []placementservice.VolumePin, currConstraints []string) *placementSettingsData {
	return &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			Service: &swarm.Service{
				Spec: swarm.ServiceSpec{
					TaskTemplate: swarm.TaskSpec{
						Placement: &swarm.Placement{Constraints: currConstraints},
					},
				},
			},
			PlacementSettings: &entity.AppPlacementSettings{},
			BuildSettings:     &entity.ImageBuildSettings{},
			VolumePins:        pins,
		},
	}
}

// A task can only run where its data is, so the pin becomes a constraint swarm
// enforces rather than a note HivePaaS keeps.
func TestApplyAddsVolumePinConstraint(t *testing.T) {
	data := applyData([]placementservice.VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}}, nil)

	(&service{}).applyPlacementSettings(data)

	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.id==node-1")
	assert.Contains(t,
		data.Service.Spec.Labels["hivepaas.app.placementConstraints"], "node.id==node-1")
	assert.True(t, data.HasChanges)
}

// Constraints the operator set by hand survive, the same way they already do for
// the exclusions HivePaaS manages.
func TestApplyKeepsOperatorConstraints(t *testing.T) {
	data := applyData(
		[]placementservice.VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}},
		[]string{"node.labels.tier==gold"},
	)

	(&service{}).applyPlacementSettings(data)

	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.labels.tier==gold")
	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.id==node-1")
}

func TestApplyAddsNothingForUnpinnedVolumes(t *testing.T) {
	data := applyData([]placementservice.VolumePin{{VolumeName: "shared"}}, nil)

	(&service{}).applyPlacementSettings(data)

	assert.Empty(t, data.Service.Spec.TaskTemplate.Placement.Constraints)
}
