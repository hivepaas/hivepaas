package imagebuildserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/docker"
)

// cmdBuildx is the docker CLI plugin every builder command goes through.
const cmdBuildx = "buildx"

// buildkitContainerPrefix is how the docker-container driver names the container it
// runs buildkit in: the prefix, the node name, and the node index. A builder created
// with --name X gets nodes X0, X1, ... so its first container is buildx_buildkit_X0.
//
// The container carries no labels at all - buildx sets none - so the name is the only
// thing that identifies it. A label filter here silently matches nothing, which looks
// exactly like a builder that has not started.
const buildkitContainerPrefix = "buildx_buildkit_"

// defaultMemSwapFactor is the total memory+swap allowance given when a memory limit
// is configured without a swap one. It is what docker itself picks for `run -m X`,
// so an operator who sets only a memory limit gets the behavior that flag describes.
// Setting memSwap equal to mem is how you ask for no swap at all.
const defaultMemSwapFactor = 2

const (
	// builderBootstrapTimeout bounds the bootstrap step, which pulls the buildkit
	// image the first time. It is also the expiry of the lock guarding that step, and
	// the two have to stay the same number: a lock that expires while the command it
	// guards is still running has stopped being a lock, which is worse than no lock at
	// all because the code still trusts it.
	builderBootstrapTimeout = 5 * time.Minute

	// builderLockTries and builderLockRetryDelay let a build wait out one full
	// bootstrap by another build. Their product has to exceed builderBootstrapTimeout,
	// otherwise the builds queued behind the first one give up before their turn and
	// fail for no reason of their own.
	builderLockTries      = 160
	builderLockRetryDelay = 2 * time.Second

	// builderRecreateCooldown stops a burst of failing builds from recreating the
	// builder once each: the first one does it, the rest see the marker and only
	// bootstrap.
	builderRecreateCooldown = 10 * time.Minute
)

// The builder is per-node state while Redis is shared by the whole cluster, so both
// keys are scoped to the node. Without that, builds on unrelated nodes would queue
// behind each other for nothing.
func builderLockKey(nodeID string) string     { return "imagebuild:builder:lock:" + nodeID }
func builderRecreateKey(nodeID string) string { return "imagebuild:builder:recreated:" + nodeID }

// ensureCustomBuilder makes sure the named buildx builder exists and that its
// buildkit container is running, then applies the configured resource limits to it.
//
// The preparation runs under a per-node lock. The lock deliberately covers only this
// step and not the build that follows: an exclusive lock held for a whole build would
// serialize every build on the node, and the lock helper has no renewal, so its
// expiry would have to be as long as the slowest build - long enough that one crashed
// worker blocks the node for that entire time.
//
// What the short lock buys is that a broken builder is recreated once rather than by
// every build at the same time. It does not remove the risk entirely: a build that
// already passed this step can still be killed by a recreate. That is bounded by the
// cooldown marker, by never recreating on our own timeout, and by the task queue
// retrying the build that lost.
func (s *service) ensureCustomBuilder(
	ctx context.Context,
	builderName string,
	res *entity.ImageBuildResourceSettings,
	logStore *tasklog.Store,
) error {
	nodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return hperrors.Wrap(err).WithMsgLog("failed to identify the current node")
	}

	// Do takes a func with no error return, so the result comes back through a capture.
	var ensureErr error
	lockErr := s.redisLock.Do(ctx, builderLockKey(nodeID), builderBootstrapTimeout,
		builderLockTries, builderLockRetryDelay,
		func() {
			ensureErr = s.ensureBuilderReady(ctx, builderName, nodeID, logStore)
		})
	if lockErr != nil {
		return hperrors.Wrap(lockErr).
			WithMsgLog("failed to acquire the builder lock on node %s", nodeID)
	}
	if ensureErr != nil {
		return hperrors.Wrap(ensureErr)
	}

	// Outside the lock: these updates are per container and idempotent, so holding the
	// lock across them would only make the other builds wait.
	s.applyBuilderResourceLimits(ctx, builderName, res, logStore)

	return nil
}

// ensureBuilderReady creates the builder if it is missing and starts its buildkit
// container, recreating the builder once when bootstrap fails in a way that points at
// the builder itself. It runs under the node's builder lock.
func (s *service) ensureBuilderReady(
	ctx context.Context,
	builderName string,
	nodeID string,
	logStore *tasklog.Store,
) error {
	if err := s.ensureBuilderExists(ctx, builderName); err != nil {
		return hperrors.Wrap(err)
	}

	timedOut, bootstrapErr := s.bootstrapBuilder(ctx, builderName, logStore)
	if bootstrapErr == nil {
		return nil
	}

	// Our own deadline running out, or the caller giving up, says nothing about the
	// health of the builder. Recreating on that would throw away a builder - and its
	// whole cache - that may be perfectly fine, and would not fix a slow image pull.
	if timedOut || ctx.Err() != nil {
		return hperrors.Wrap(bootstrapErr)
	}
	// Another build recreated it moments ago and it still will not start. Doing it
	// again would only repeat their failure and destroy the cache a second time.
	if !s.claimBuilderRecreate(ctx, nodeID) {
		return hperrors.Wrap(bootstrapErr)
	}

	logging.Warnf("image build: recreating builder %q on node %s: %v", builderName, nodeID, bootstrapErr)
	_ = logStore.Add(ctx, tasklog.NewWarnFrame(
		"The build engine failed to start; recreating it (the build cache will be lost)...", tasklog.TsNow))

	if out, err := runBuildx(ctx, "rm", "--force", builderName); err != nil {
		// Report the bootstrap failure, not this one: the removal failing is a symptom.
		return hperrors.Wrap(bootstrapErr).
			WithMsgLog("failed to remove the broken builder: %s", reflectutil.UnsafeBytesToStr(out))
	}
	if err := s.ensureBuilderExists(ctx, builderName); err != nil {
		return hperrors.Wrap(err)
	}

	_, err := s.bootstrapBuilder(ctx, builderName, logStore)
	return err
}

// bootstrapBuilder starts the buildkit container, pulling its image the first time.
// That makes it the slowest step here and the one most likely to fail, so it is
// announced rather than leaving the build log silent while it runs.
//
// timedOut reports that the step hit its own deadline rather than failing outright,
// which the caller needs in order not to treat a slow network as a broken builder.
func (s *service) bootstrapBuilder(
	ctx context.Context,
	builderName string,
	logStore *tasklog.Store,
) (timedOut bool, err error) {
	_ = logStore.Add(ctx, tasklog.NewOutFrame(
		"Preparing the build engine (may pull the buildkit image on first use)...", tasklog.TsNow))

	// The deadline matches the lock expiry so the command cannot outlive the lock.
	bootstrapCtx, cancel := context.WithTimeout(ctx, builderBootstrapTimeout)
	defer cancel()

	out, err := runBuildx(bootstrapCtx, "inspect", "--bootstrap", builderName)
	if err != nil {
		return errors.Is(bootstrapCtx.Err(), context.DeadlineExceeded),
			hperrors.Wrap(err).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
	}
	return false, nil
}

// claimBuilderRecreate takes the right to recreate the builder on this node, and
// reports false when another build already took it within the cooldown. The lock only
// keeps recreates from overlapping; this is what keeps them from repeating.
func (s *service) claimBuilderRecreate(ctx context.Context, nodeID string) bool {
	claimed, err := s.redisClient.
		SetNX(ctx, builderRecreateKey(nodeID), time.Now().Unix(), builderRecreateCooldown).Result()
	if err != nil {
		// Reaching here means Redis answered for the lock and not for this, so it just
		// became unreachable. Allow the recreate: the lock already holds it to one at a
		// time, and refusing would leave a broken builder broken.
		logging.Warnf("image build: failed to claim the builder recreate marker: %v", err)
		return true
	}
	return claimed
}

// ensureBuilderExists creates the builder when it is missing.
//
// `buildx inspect` failing does not by itself mean the builder is missing: a stopped
// daemon, a docker CLI that is not on PATH and a canceled context all fail exactly
// the same way. Creating a builder in those cases only buries the real cause under a
// second, more confusing failure, so docker is probed first and its own error is
// reported as itself.
//
// The builder is shared by every build on the node (see base.HivepaasGlobalBuilder)
// while builds run concurrently, so two of them can reach the create step together.
// One will lose and be told the instance already exists; that is a success, not a
// failure - what matters is that the builder is there afterwards, not who made it.
func (s *service) ensureBuilderExists(ctx context.Context, builderName string) error {
	if _, err := runBuildx(ctx, "inspect", builderName); err == nil {
		return nil
	}

	if out, err := runDocker(ctx, "version", "--format", "{{.Server.Version}}"); err != nil {
		return hperrors.Wrap(err).
			WithMsgLog("docker is not usable: %s", reflectutil.UnsafeBytesToStr(out))
	}

	out, createErr := runBuildx(ctx, "create", "--name", builderName, "--driver", "docker-container")
	if createErr == nil {
		return nil
	}
	if _, err := runBuildx(ctx, "inspect", builderName); err == nil {
		return nil
	}

	return hperrors.Wrap(createErr).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
}

// applyBuilderResourceLimits pushes the configured CPU and memory limits onto the
// buildkit container.
//
// A failure here does not fail the build: the builder works, it just runs
// unconstrained. It does get reported though, because a limit that is silently not
// applied is indistinguishable from a limit that does not work, and the symptom
// surfaces much later as a build that eats the node.
func (s *service) applyBuilderResourceLimits(
	ctx context.Context,
	builderName string,
	res *entity.ImageBuildResourceSettings,
	logStore *tasklog.Store,
) {
	hasLimits := res != nil && (res.CPUs > 0 || res.Mem > 0 || res.MemSwap > 0)
	if !hasLimits {
		return
	}

	resList, err := s.dockerManager.ContainerList(ctx, func(opts *client.ContainerListOptions) {
		opts.All = true
		docker.FilterAdd(&opts.Filters, "name", buildkitContainerPattern(builderName))
	})
	if err != nil {
		warnBuilderLimits(ctx, logStore, "failed to list the containers of builder %q: %v", builderName, err)
		return
	}
	if resList == nil || len(resList.Items) == 0 {
		warnBuilderLimits(ctx, logStore, "no container found for builder %q", builderName)
		return
	}

	updateRes, skipped := buildResourceUpdate(res)
	for _, reason := range skipped {
		warnBuilderLimits(ctx, logStore, "%s", reason)
	}

	for _, item := range resList.Items {
		_, err := s.dockerManager.ContainerUpdate(ctx, item.ID, func(opts *client.ContainerUpdateOptions) {
			opts.Resources = &updateRes
		})
		if err != nil {
			warnBuilderLimits(ctx, logStore, "failed to update container %s: %v", item.ID, err)
		}
	}
}

// buildResourceUpdate turns the configured limits into a container update, together
// with the reason for any limit it had to leave out.
//
// Memory and swap have to travel together. Docker refuses a memory update that is
// larger than the swap limit already on the container unless the same call sets swap
// too, and the buildkit container is created with neither - so sending a memory limit
// on its own always fails with "update the memoryswap at the same time".
func buildResourceUpdate(res *entity.ImageBuildResourceSettings) (container.Resources, []string) {
	var update container.Resources
	var skipped []string

	if res.CPUs > 0 {
		update.NanoCPUs = int64(res.CPUs) * docker.UnitCPUNano //nolint:gosec
	}

	switch {
	case res.Mem > 0 && res.MemSwap > 0 && res.MemSwap < res.Mem:
		// Docker would reject the pair. Report the numbers rather than quietly
		// raising one of them: which one the operator meant is not ours to guess.
		skipped = append(skipped, fmt.Sprintf(
			"memory limits not applied: memSwap (%s) must be at least mem (%s)", res.MemSwap, res.Mem))

	case res.Mem > 0:
		update.Memory = res.Mem.Bytes()
		if res.MemSwap > 0 {
			update.MemorySwap = res.MemSwap.Bytes()
		} else {
			update.MemorySwap = res.Mem.Bytes() * defaultMemSwapFactor
		}

	case res.MemSwap > 0:
		skipped = append(skipped, fmt.Sprintf(
			"memSwap (%s) not applied: docker requires a mem limit alongside it", res.MemSwap))
	}

	return update, skipped
}

func warnBuilderLimits(ctx context.Context, logStore *tasklog.Store, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logging.Warnf("image build: resource limits not applied: %s", msg)
	_ = logStore.Add(ctx, tasklog.NewWarnFrame("Build resource limits not applied: "+msg, tasklog.TsNow))
}

// buildkitContainerPattern matches every buildkit container of one builder.
//
// The docker name filter is a regular expression matched against the stored names,
// which carry a leading slash. It is anchored so a builder named "web" does not also
// match the containers of "web-staging", and the name is quoted because it comes from
// configuration rather than from this package.
func buildkitContainerPattern(builderName string) string {
	return fmt.Sprintf("^/?%s%s[0-9]+$", buildkitContainerPrefix, regexp.QuoteMeta(builderName))
}

func runDocker(ctx context.Context, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return out, hperrors.Wrap(err)
	}
	return out, nil
}

func runBuildx(ctx context.Context, args ...string) ([]byte, error) {
	return runDocker(ctx, append([]string{cmdBuildx}, args...)...)
}
