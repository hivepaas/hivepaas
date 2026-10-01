// Package functioncontainer is what is fixed for a function's container,
// whatever a deployment or a screen would otherwise set.
package functioncontainer

import (
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// StopGraceMargin is how much longer than one call's timeout a stopping
// function is given: the runtime lets the running calls finish, for at most one
// timeout, and needs a moment to exit after them.
const StopGraceMargin = 10 * time.Second

// ApplyFixed gives a function's container what is fixed for a function: the
// runtime's command, working directory and health check, whatever a setting or
// an earlier image left; and a stop grace period that lets the calls running
// at a stop finish.
func ApplyFixed(contSpec *swarm.ContainerSpec, source *entity.DeploymentFunctionSource) {
	contSpec.Dir = ""
	dockerhelper.ContainerCommandApply(contSpec, "")
	contSpec.Healthcheck = nil
	grace := time.Duration(source.Timeout) + StopGraceMargin
	contSpec.StopGracePeriod = &grace
}
