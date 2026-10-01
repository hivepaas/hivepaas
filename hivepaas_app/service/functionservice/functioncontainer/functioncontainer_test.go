package functioncontainer

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// A function's container is the runtime's: its command, its working directory
// and its health check, whatever a setting or an earlier image left; and a stop
// lets a call that is running finish.
func TestAFunctionsContainerIsItsRuntimes(t *testing.T) {
	grace := time.Second
	contSpec := &swarm.ContainerSpec{
		Command: []string{"node"}, Args: []string{"server.js"}, Dir: "/srv",
		Healthcheck:     &container.HealthConfig{Test: []string{"CMD", "true"}},
		StopGracePeriod: &grace,
	}

	ApplyFixed(contSpec, &entity.DeploymentFunctionSource{Timeout: timeutil.Duration(30 * time.Second)})

	assert.Nil(t, contSpec.Command)
	assert.Nil(t, contSpec.Args)
	assert.Empty(t, contSpec.Dir)
	assert.Nil(t, contSpec.Healthcheck)
	assert.Equal(t, 40*time.Second, *contSpec.StopGracePeriod)
}
