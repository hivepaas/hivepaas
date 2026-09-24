package sysupdateserviceimpl

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// pullConcurrency caps how many images are fetched at once.
	pullConcurrency = 3

	// pullRetryMax and pullRetryDelay cover a blip, not an outage. Three attempts
	// spaced a few seconds apart is the shape of "the registry hiccupped"; longer
	// than that and the update is better off letting swarm fetch the image when
	// it starts the task.
	pullRetryMax   = 2
	pullRetryDelay = 5 * time.Second
)

// pullAllImages fetches everything the release names, before anything is stopped.
//
// This runs outside the window where the system is down, which is the whole
// point of where it is called from: pulling is the one part of an update whose
// cost has nothing to do with the system being unavailable, and there is no
// reason to pay it with the app switched off. On a slow link that is minutes.
//
// It also turns an unreachable registry or a mistyped tag into a refusal before
// the first service is touched, rather than a failure halfway through with the
// app already at zero replicas.
//
// A missing image is not fatal here. Swarm pulls what it needs when it starts a
// task, so a failure means the update is slower and noisier, not broken - which
// is why the caller logs the error and carries on.
func (s *service) pullAllImages(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	images := []string{
		args.TargetVersion.AppImage,
		args.TargetVersion.RedisImage,
		args.TargetVersion.DbImage,
		args.TargetVersion.TraefikImage,
		args.TargetVersion.VictoriaLogsImage,
		args.TargetVersion.VlagentImage,
		args.TargetVersion.RegistryImage,
	}

	tasks := make([]func(context.Context) error, 0, len(images))
	for _, image := range images {
		tasks = append(tasks, func(ctx context.Context) error {
			return s.pullImage(ctx, image, data)
		})
	}

	// Three at a time rather than all of them. Every image at once would contend for
	// the same bandwidth and disk, and against an unauthenticated registry they
	// are also that many requests toward a rate limit - which no amount of retrying
	// gets past, so the cheaper fix is not to approach it.
	//
	// Not stopping at the first error: one image being unreachable is no reason
	// to leave the others unpulled, and every failure is worth reporting.
	errMap := gofn.ExecTasksEx(ctx, pullConcurrency, false, tasks...)

	pullErrs := make([]error, 0, len(errMap))
	for _, err := range errMap {
		pullErrs = append(pullErrs, err)
	}
	return errors.Join(pullErrs...)
}

// pullImage fetches one image, retrying a failure a couple of times.
//
// What the retry is for is narrow and worth being clear about: a blip - a 5xx, a
// dropped connection, a name that did not resolve for a moment. It does not get
// past a registry rate limit, which lasts hours; pullConcurrency is what keeps
// that from being reached in the first place.
//
// Retrying costs little because this runs before anything is stopped. A tag that
// does not exist is attempted three times and then reported, and the seconds
// that wastes are seconds the system is still serving - which is why no effort
// goes into telling that case apart from a blip.
func (s *service) pullImage(
	ctx context.Context,
	image string,
	data *sysUpdateData,
) (err error) {
	if image == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image, tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image+" finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image+" finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	attempt := 0
	err = gofn.ExecRetryCtx(ctx, func() error {
		attempt++
		if attempt > 1 {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
				"Retrying the pull of "+image+" (attempt "+strconv.Itoa(attempt)+")", tasklog.TsNow))
		}
		return s.pullImageOnce(ctx, image, data)
	}, pullRetryMax, pullRetryDelay)
	return hperrors.Wrap(err)
}

// pullImageOnce is one attempt, and the whole of it is retryable.
//
// Both of this function's failures have to be inside the retry, which is why it
// exists: ImagePull returns one kind of error, and the stream it hands back
// carries the other in its own messages. Wrapping only the call would retry half
// the failures and report the rest on the first try.
func (s *service) pullImageOnce(
	ctx context.Context,
	image string,
	data *sysUpdateData,
) (err error) {
	logsReader, err := s.dockerManager.ImagePull(ctx, image)
	if err != nil {
		return hperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := tasklog.NewOutFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = tasklog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.Add(ctx, frameCreator(msg.String(), tasklog.TsNow))
			}
		}
	}
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
