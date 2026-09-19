package appdeploymentserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// applyContainerInit decides whether the container is given docker's init, for
// an app whose settings have not decided it themselves.
//
// An app is created without an answer, because the answer is the image's: one
// that starts with an init of its own keeps it, and one that starts with the
// app itself is given docker's, so that what the app orphans is still reaped.
// Two inits in one container is what makes tini warn that it cannot reap from
// where it is, and what makes s6 refuse to start at all.
//
// It is decided here rather than when the app is created because this is where
// the image is: it has just been pulled, or just been built. Once decided it is
// written down, so an app keeps the answer it was deployed with until somebody
// changes it in its container settings - where an empty value asks for this
// again on the next deployment.
func (s *service) applyContainerInit(
	ctx context.Context,
	data *appDeploymentData,
	contSpec *swarm.ContainerSpec,
) {
	if contSpec.Init != nil {
		return
	}

	dockerInit := true
	inspect, err := s.dockerManager.ImageInspect(ctx, contSpec.Image)
	switch {
	case err != nil:
		// The image is there - it was just pulled - so this is something else
		// going wrong, and the app still has to deploy. Docker's init is what
		// every app had before there was a choice.
		_ = data.LogStore.Add(ctx, tasklog.NewDebugFrame(
			"Could not read the image's entry point, keeping the init process: "+err.Error(), tasklog.TsNow))
	case inspect.Config != nil && dockerhelper.ImageProvidesInit(inspect.Config.Entrypoint):
		dockerInit = false
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"The image starts with an init of its own; leaving it as the container's", tasklog.TsNow))
	}
	contSpec.Init = &dockerInit
}
