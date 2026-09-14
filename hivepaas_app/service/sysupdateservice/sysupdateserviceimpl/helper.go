package sysupdateserviceimpl

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	// serviceUpdateCheckInterval is how often swarm is asked whether an update
	// has converged. One value for every service: they were separate constants
	// with the same value, which only made them look meaningful.
	serviceUpdateCheckInterval = time.Second * 5

	// updateMaxFailureRatio is the share of tasks allowed to fail before swarm
	// calls the update failed and acts on FailureAction.
	updateMaxFailureRatio = 0.5

	// serviceUpdateRetryMax is how many times a rejected service update is
	// retried, matching what the rest of the codebase uses.
	//
	// The retry is worth more than it looks. ServiceUpdateFunc re-inspects the
	// service before each attempt, so the change is re-applied to a fresh
	// Version - and a stale Version is exactly what swarm rejects when something
	// else has written to the service in between. Retrying the original call
	// would fail the same way every time.
	serviceUpdateRetryMax = 2
)

// scaleServiceReplicas sets a service's replica count.
//
// It matters more than its size. Two of its four callers are stopServices, which
// runs while the main app is still alive and rewriting its own traefik labels -
// a concurrent writer on the very service being scaled - and onAfterSystemUpdate,
// which is what brings the app back when an update ends. The second failing is
// how an installation is left stopped with nothing running that could fix it.
//
// The count is compared inside the callback rather than before it, so the answer
// comes from the service as it is now and not from whatever the caller fetched
// earlier.
func (s *service) scaleServiceReplicas(
	ctx context.Context,
	service *swarm.Service,
	replicas uint64,
) error {
	err := s.dockerManager.ServiceUpdateFunc(ctx, service.ID, service,
		func(_ int, svc *swarm.Service) (bool, error) {
			if svc.Spec.Mode.Replicated == nil {
				return false, nil
			}
			if svc.Spec.Mode.Replicated.Replicas != nil && *svc.Spec.Mode.Replicated.Replicas == replicas {
				return false, nil
			}
			svc.Spec.Mode.Replicated.Replicas = &replicas
			return true, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// serviceImageUpdate is one step of a system update: a service, the image the
// release wants it on, and whatever else about its spec has to move with it.
type serviceImageUpdate struct {
	// What names the service the way the update log should. It is the operator's
	// word for it - "redis", "hivepaas worker" - not the swarm service name.
	What string

	// Component is the key the release uses for this service - HivepaasDbKey and
	// friends - and is what BlockMajorUpgrade is matched against. Leaving it empty
	// means the release has no say over this service's major version.
	Component string

	// TargetImage is what the release says to run. Empty means this step has
	// nothing to do and is not an error: a release need not name every image.
	TargetImage string

	// Fetch returns the service to update. A nil service with a nil error means
	// it is not deployed, which for the optional ones - the worker, the logging
	// stack - is an ordinary state and not a failure.
	Fetch func(ctx context.Context) (*swarm.Service, error)

	// Mutate applies whatever else the service needs, and is called only once the
	// image has been found worth moving to. The image itself is already set.
	Mutate func(spec *swarm.ServiceSpec)
}

// updateServiceImage runs one step: announce it, decide whether the image is
// worth moving to, apply it, and wait for swarm to say whether it stuck.
//
// Deciding not to move an image is a success, not a no-op the caller has to
// handle: a step that finds the service already current has done its job. What
// still has to happen either way - the database migrations, the replicas coming
// back - is the caller's, and does not depend on the answer.
func (s *service) updateServiceImage(
	ctx context.Context,
	data *sysUpdateData,
	step serviceImageUpdate,
) (err error) {
	if step.TargetImage == "" {
		return nil
	}

	defer s.logStep(ctx, data, step.What)(&err)

	svc, err := step.Fetch(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if svc == nil {
		_ = data.LogStore.Add(ctx,
			tasklog.NewOutFrame("Skipping "+step.What+": not deployed", tasklog.TsNow))
		return nil
	}

	// Decided inside the callback so that a retry judges the service as it is
	// now. If the first attempt actually landed and only its answer was lost,
	// the next one reads the image it was moving to and stops - so a retry can
	// never roll the same service out twice.
	applied := false
	err = s.dockerManager.ServiceUpdateFunc(ctx, svc.ID, svc,
		func(_ int, current *swarm.Service) (bool, error) {
			currentImage := current.Spec.TaskTemplate.ContainerSpec.Image
			if !s.shouldUpdateImage(ctx, data, step.What, currentImage, step.TargetImage) {
				return false, nil
			}
			// Checked only once the image has been found worth moving to. An
			// update that was going to change nothing has no major to cross.
			if err := s.checkMajorUpgrade(ctx, data, step, currentImage); err != nil {
				return false, hperrors.Wrap(err)
			}

			current.Spec.TaskTemplate.ContainerSpec.Image = step.TargetImage
			if step.Mutate != nil {
				step.Mutate(&current.Spec)
			}

			// Every service the updater touches gets this. Without it swarm
			// applies an update that brings the service down and leaves it down,
			// and the operator is left to notice and undo it by hand.
			if current.Spec.UpdateConfig == nil {
				current.Spec.UpdateConfig = &swarm.UpdateConfig{}
			}
			current.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
			current.Spec.UpdateConfig.MaxFailureRatio = updateMaxFailureRatio

			applied = true
			return true, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Nothing was sent, so there is no rollout of ours to wait on.
	if !applied {
		return nil
	}

	updated, err := s.dockerManager.ServiceUpdateWait(ctx, svc.ID, serviceUpdateCheckInterval)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Swarm undid it on its own. The service is still up, on the old image, and
	// the update as a whole has not done what it was asked.
	if updated.UpdateStatus != nil && updated.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx,
			tasklog.NewWarnFrame("service "+step.What+" is rolled back", tasklog.TsNow))
		return hperrors.Wrap(hperrors.ErrActionFailed)
	}

	return nil
}

// logStep records a step starting, and returns what to defer so that the same
// line reports how long it took and whether it failed.
//
// It takes the address of the caller's error rather than its value, because a
// deferred call's arguments are evaluated when it is deferred - at which point
// the step has not run yet and the error is always nil.
func (s *service) logStep(ctx context.Context, data *sysUpdateData, what string) func(err *error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating "+what+"...", tasklog.TsNow))

	return func(err *error) {
		duration := timeutil.NowUTC().Sub(start).String()
		if err != nil && *err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating "+what+" finished in "+duration+
				" with error: "+(*err).Error(), tasklog.TsNow))
			return
		}
		_ = data.LogStore.Add(ctx,
			tasklog.NewOutFrame("Updating "+what+" finished in "+duration, tasklog.TsNow))
	}
}

// shouldUpdateImage answers whether a step should move a service to target, and
// records the reason in the task log either way.
//
// Applying an image that is not newer is not free: every swarm service update
// restarts the task, so re-running an update - after a failure, or because the
// same release was applied twice - would otherwise restart the whole system for
// nothing.
//
// The reason is logged for a refusal as well as for a go-ahead, because a step
// that decided to do nothing otherwise reads exactly like a step that was never
// reached, and the two call for different reactions from whoever is watching.
//
// `current` comes from the running service's own spec rather than from the
// task's CurrentVersion: that field is filled from the constants compiled into
// whichever binary created the task, which is a claim about the release, not an
// observation of what is deployed.
func (s *service) shouldUpdateImage(
	ctx context.Context,
	data *sysUpdateData,
	what string,
	current string,
	target string,
) bool {
	apply, reason := imageref.IsUpgrade(current, target)
	if !apply {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Skipping "+what+": "+reason, tasklog.TsNow))
		return false
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(what+": "+reason, tasklog.TsNow))
	return true
}

// checkMajorUpgrade refuses a step that would cross a major version the release
// says must not be crossed.
//
// The decision belongs to the release, not to this code and not to the operator:
// by the time a release names a new major, whoever cut it has read the upstream
// notes, and neither of the other two ever could. See ReleaseInfo.
//
// A tag with no version it can read is let through, deliberately. Blocking every
// release that names an image this cannot parse would be the more expensive
// mistake, and the check exists for one specific event, not as a gate on all
// image changes.
func (s *service) checkMajorUpgrade(
	ctx context.Context,
	data *sysUpdateData,
	step serviceImageUpdate,
	currentImage string,
) error {
	if step.Component == "" {
		return nil
	}
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
	if !slices.Contains(args.TargetVersion.BlockMajorUpgrade, step.Component) {
		return nil
	}

	current, currentOK := imageref.MajorVersion(currentImage)
	target, targetOK := imageref.MajorVersion(step.TargetImage)
	if !currentOK || !targetOK || current == target {
		return nil
	}

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Refusing to update "+step.What+": this release moves it from major version "+
			strconv.Itoa(current)+" to "+strconv.Itoa(target)+
			", which needs a data migration this update does not perform",
		tasklog.TsNow))
	return hperrors.Wrap(hperrors.ErrUnsupported).
		WithMsgLog("%s would move %s from major %d to %d", step.TargetImage, step.Component, current, target)
}
