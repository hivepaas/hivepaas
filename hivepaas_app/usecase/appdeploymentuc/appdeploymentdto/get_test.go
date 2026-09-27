package appdeploymentdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A deployment answers the change it was made for, as its trigger gave it.
func TestTransformDeploymentAnswersTheChangeID(t *testing.T) {
	resp, err := TransformDeployment(&entity.Deployment{
		ID:       "d1",
		Settings: &entity.AppDeploymentSettings{},
		Trigger:  &entity.AppDeploymentTrigger{Source: base.DeploymentTriggerSourceAPI, SourceID: "u1", ChangeID: "pr-12"},
	}, &DeploymentTransformInput{})
	if assert.NoError(t, err) && assert.NotNil(t, resp.Trigger) {
		assert.Equal(t, "pr-12", resp.Trigger.ChangeID)
	}
}
