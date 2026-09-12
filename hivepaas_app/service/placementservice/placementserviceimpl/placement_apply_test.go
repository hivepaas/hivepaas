package placementserviceimpl

import (
	"fmt"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
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

// warnRecordingLogger keeps only the warnings, which is all this branch emits.
type warnRecordingLogger struct {
	logging.Logger

	warnings []string
}

func (l *warnRecordingLogger) Warnf(template string, args ...any) {
	l.warnings = append(l.warnings, fmt.Sprintf(template, args...))
}

// Two volumes pinned to different nodes describe a placement no node satisfies,
// so no constraint can be emitted. UpdateAppStorageSettings refuses that set
// before it is saved, but a spec written before that check existed - or edited
// on the service directly - still reaches here, and a service that has silently
// lost its pin has to be visible somewhere.
func TestApplyWarnsWhenPinsConflictInsteadOfSilentlyDroppingTheConstraint(t *testing.T) {
	data := applyData([]placementservice.VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-2"},
	}, nil)
	data.Service.Spec.Name = "shop-prod-web"
	logger := &warnRecordingLogger{}

	(&service{logger: logger}).applyPlacementSettings(data)

	assert.Empty(t, data.Service.Spec.TaskTemplate.Placement.Constraints)
	if len(logger.warnings) != 1 {
		t.Fatalf("expected exactly 1 warning, got %d: %v", len(logger.warnings), logger.warnings)
	}
	assert.Contains(t, logger.warnings[0], "shop-prod-web")
	assert.Contains(t, logger.warnings[0], "pgdata")
	assert.Contains(t, logger.warnings[0], "uploads")
}

// The operator's label rules reach swarm as constraints, and are recorded as
// HivePaaS-set so that a later change can take them back out.
func TestApplyAddsNodeLabelRules(t *testing.T) {
	data := applyData(nil, nil)
	data.PlacementSettings.RequireNodeLabels = []string{"zone=eu", "disk"}
	data.PlacementSettings.ExcludeNodeLabels = []string{"maintenance=true"}

	(&service{}).applyPlacementSettings(data)

	constraints := data.Service.Spec.TaskTemplate.Placement.Constraints
	assert.Contains(t, constraints, "node.labels.zone==eu")
	assert.Contains(t, constraints, "node.labels.disk==true")
	assert.Contains(t, constraints, "node.labels.maintenance!=true")
	recorded := data.Service.Spec.Labels["hivepaas.app.placementConstraints"]
	assert.Contains(t, recorded, "node.labels.zone==eu")
	assert.Contains(t, recorded, "node.labels.maintenance!=true")
	assert.True(t, data.HasChanges)
}

// Rules that are removed from the settings must leave the service with them.
func TestApplyRemovesNodeLabelRulesItAddedBefore(t *testing.T) {
	data := applyData(nil, []string{"node.labels.zone==eu", "node.labels.keep==mine"})
	data.Service.Spec.Labels = map[string]string{
		"hivepaas.app.placementConstraints": "node.labels.zone==eu",
	}

	(&service{}).applyPlacementSettings(data)

	constraints := data.Service.Spec.TaskTemplate.Placement.Constraints
	assert.NotContains(t, constraints, "node.labels.zone==eu")
	assert.Contains(t, constraints, "node.labels.keep==mine", "the operator set this one by hand")
}
