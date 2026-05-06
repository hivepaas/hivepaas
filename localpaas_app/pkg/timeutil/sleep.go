package timeutil

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/pkg/tracerr"
)

func SleepCtx(
	ctx context.Context,
	sleepDuration time.Duration,
) error {
	if sleepDuration <= 0 {
		if err := ctx.Err(); err != nil {
			return tracerr.Wrap(err)
		}
		return nil
	}

	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return tracerr.Wrap(ctx.Err())
	case <-timer.C:
		return nil
	}
}
