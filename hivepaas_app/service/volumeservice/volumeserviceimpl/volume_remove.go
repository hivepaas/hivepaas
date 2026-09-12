package volumeserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// Docker counts a volume as in use while any container still references it,
	// a stopped one included, and swarm shuts task containers down in the
	// background - so a removal issued right after the service that used it was
	// removed loses a race the same call would win a moment later.
	//
	// How long that race lasts was measured rather than guessed: the volume
	// becomes removable about two seconds after the task container actually
	// exits, and the container does not exit until it has either handled SIGTERM
	// or sat out its whole stop-grace-period. Against a service with the default
	// ten second grace running a process that ignores SIGTERM, that came to
	// 11.9s; at a thirty second grace, 32.0s.
	//
	// Four retries at 250ms growing by 500ms is four seconds of waiting across
	// five calls, which deliberately does not cover those numbers. Volume
	// removal runs inside the database transaction that deletes the volume's
	// setting, and no budget that would cover a ten second grace period is one
	// worth holding a transaction open for. What this does cover is the far more
	// common case where the task is already gone - a volume detached earlier, an
	// app deleted a while ago, a container that exits promptly on SIGTERM - and
	// only the last couple of seconds of reaping stand in the way. A volume
	// whose task is still shutting down reports the conflict to the caller
	// instead, and deleting it again a moment later succeeds.
	volumeRemovalRetryMax   = 4
	volumeRemovalRetryDelay = 250 * time.Millisecond
	volumeRemovalRetryIncr  = 500 * time.Millisecond
)

// RemoveVolume removes a volume from docker, waiting out the window in which it
// is still counted as in use.
//
// Only that conflict is retried. Every other failure is returned on the first
// attempt, because none of them turn into success by waiting. A volume that is
// already gone is success: force does not make a missing volume an error, and
// neither does this.
//
// Passing retryMax <= 0 removes without retrying, for callers that would rather
// fail fast than hold a caller waiting.
func (s *service) RemoveVolume(
	ctx context.Context,
	volumeID string,
	force bool,
	retryMax int,
	retryDelay time.Duration,
) error {
	if volumeID == "" {
		return nil
	}

	fn := func() error {
		_, err := s.dockerManager.VolumeRemove(ctx, volumeID, force)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				return nil
			}
			return hperrors.Wrap(err)
		}
		return nil
	}

	var err error
	if retryMax > 0 {
		if retryDelay <= 0 {
			retryDelay = volumeRemovalRetryDelay
		}
		err = gofn.ExecRetryCtx(ctx, fn, retryMax, retryDelay,
			gofn.ExecRetryDelayIncr(volumeRemovalRetryIncr),
			// Narrower than the siblings, which retry anything: a conflict is
			// the one failure here that a short wait actually resolves.
			gofn.ExecRetryIfErrorIs(hperrors.ErrConflict),
		)
	} else {
		err = fn()
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
