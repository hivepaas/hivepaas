package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// What one deployment asked for travels with its task rather than with the app's
// settings, so it has to survive being written and read back.
func TestDeployArgsCarryWhatOneDeploymentAskedFor(t *testing.T) {
	task := &entity.Task{ID: "t1"}
	task.MustSetArgs(&entity.TaskAppDeployArgs{
		Deployment: entity.ObjectID{ID: "d1"},
		NoCache:    true,
		ImageTags:  []string{"v1.4.0"},
	})

	// Read it the way the executor does: from the stored string, not the cache.
	stored := &entity.Task{ID: task.ID, Args: task.Args}
	args, err := stored.ArgsAsAppDeploy()

	assert.NoError(t, err)
	assert.Equal(t, "d1", args.Deployment.ID)
	assert.True(t, args.NoCache)
	assert.Equal(t, []string{"v1.4.0"}, args.ImageTags)
}
