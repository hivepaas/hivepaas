package imagescanserviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// busyVolume is a volume another scan holds: every try fails; it keeps how
// many it was given.
type busyVolume struct{ tries int }

func (l *busyVolume) Do(_ context.Context, _ string, _ time.Duration, tries int, _ time.Duration, _ func(),
	_ ...redsync.Option) error {
	l.tries = tries
	return errors.New("lock already taken")
}

// A scan waits for the volume as long as a scan may take; an SBOM, which a
// person waits on, half a minute - and says to try again.
func TestAnSBOMWaitsForTheVolumeHalfAMinute(t *testing.T) {
	lock := &busyVolume{}
	svc := &service{redisLock: lock}
	volume := &scanVolume{VolumeMount: &volumeservice.VolumeMount{Setting: &entity.Setting{ID: "vol"}}}

	_, err := svc.runLocked(context.Background(), volume, "n1", sbomLockTries, &imagescanrun.RunReq{})

	assert.ErrorIs(t, err, hperrors.ErrImageScanNotReady)
	assert.Equal(t, 15, lock.tries)
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "did not end within 30s; try again later")

	_, _ = svc.runLocked(context.Background(), volume, "n1", lockTries, &imagescanrun.RunReq{})
	assert.Equal(t, int(callTimeout/lockRetryDelay), lock.tries, "a scan waits a scan's length")
}
