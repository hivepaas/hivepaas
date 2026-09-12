package volumeserviceimpl

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/containerd/errdefs/pkg/errhttp"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// testRetryDelay only has to be non-zero. The delay that matters for test
// runtime is the first one, which RemoveVolume takes from this argument;
// volumeRemovalRetryIncr only widens later ones, so every case here keeps
// retryMax low enough that no later delay is ever waited out.
const testRetryDelay = time.Millisecond

// dockerConflictErr rebuilds the error a docker 409 has become by the time
// RemoveVolume sees it: the moby client maps the status code through
// errhttp.ToNative, and services/docker runs every client error through
// hperrors.NewInfra.
func dockerConflictErr() error {
	return hperrors.NewInfra(errhttp.ToNative(http.StatusConflict))
}

func dockerNotFoundErr() error {
	return hperrors.NewInfra(errhttp.ToNative(http.StatusNotFound))
}

// fakeRemoveManager answers a scripted sequence of results, one per call, and
// repeats the last one once the script runs out.
type fakeRemoveManager struct {
	docker.Manager
	results []error
	calls   int
	lastID  string
	lastFrc bool
}

func (f *fakeRemoveManager) VolumeRemove(
	_ context.Context, volumeID string, force bool, _ ...docker.VolumeRemoveOption,
) (*client.VolumeRemoveResult, error) {
	f.calls++
	f.lastID = volumeID
	f.lastFrc = force

	i := f.calls - 1
	if i >= len(f.results) {
		i = len(f.results) - 1
	}
	if i < 0 || f.results[i] == nil {
		return &client.VolumeRemoveResult{}, nil
	}
	return nil, f.results[i]
}

func newRemoveTest(results ...error) (*service, *fakeRemoveManager) {
	mgr := &fakeRemoveManager{results: results}
	return &service{dockerManager: mgr}, mgr
}

// Everything else here rests on this: if a docker 409 did not arrive as an
// ErrConflict, gofn.ExecRetryIfErrorIs would never fire and RemoveVolume would
// quietly stop retrying.
func TestDockerConflictIsErrConflict(t *testing.T) {
	assert.True(t, errors.Is(dockerConflictErr(), hperrors.ErrConflict))
	assert.False(t, errors.Is(dockerConflictErr(), hperrors.ErrNotFound))
	assert.True(t, errors.Is(dockerNotFoundErr(), hperrors.ErrNotFound))
	assert.False(t, errors.Is(dockerNotFoundErr(), hperrors.ErrConflict))
}

func TestRemoveVolumeRetriesUntilTheVolumeIsFree(t *testing.T) {
	svc, mgr := newRemoveTest(dockerConflictErr(), dockerConflictErr(), nil)

	err := svc.RemoveVolume(context.Background(), "vol-1", true, 3, testRetryDelay)

	assert.NoError(t, err)
	assert.Equal(t, 3, mgr.calls)
	assert.Equal(t, "vol-1", mgr.lastID)
	assert.True(t, mgr.lastFrc)
}

// Waiting turns a conflict into a success. Nothing else, so nothing else should
// cost the caller a delay.
func TestRemoveVolumeDoesNotRetryOtherErrors(t *testing.T) {
	boom := hperrors.NewInfra(errhttp.ToNative(http.StatusInternalServerError))
	svc, mgr := newRemoveTest(boom)

	err := svc.RemoveVolume(context.Background(), "vol-1", true, 4, testRetryDelay)

	assert.Equal(t, 1, mgr.calls)
	assert.Error(t, err)
}

// A volume genuinely still in use has to surface as an error the caller can
// report, not as a silent success.
func TestRemoveVolumeGivesUpAndReturnsTheConflict(t *testing.T) {
	svc, mgr := newRemoveTest(dockerConflictErr())

	err := svc.RemoveVolume(context.Background(), "vol-1", true, 1, testRetryDelay)

	assert.Equal(t, 2, mgr.calls) // the first attempt plus the one retry
	assert.True(t, errors.Is(err, hperrors.ErrConflict))
}

// The normal case after lazy materialization: the volume was never created on
// this node, so there is nothing here to remove.
func TestRemoveVolumeTreatsMissingVolumeAsDone(t *testing.T) {
	svc, mgr := newRemoveTest(dockerNotFoundErr())

	err := svc.RemoveVolume(context.Background(), "vol-1", true, 4, testRetryDelay)

	assert.NoError(t, err)
	assert.Equal(t, 1, mgr.calls)
}

func TestRemoveVolumeWithoutRetryCallsDockerOnce(t *testing.T) {
	svc, mgr := newRemoveTest(dockerConflictErr())

	err := svc.RemoveVolume(context.Background(), "vol-1", true, 0, testRetryDelay)

	assert.Equal(t, 1, mgr.calls)
	assert.True(t, errors.Is(err, hperrors.ErrConflict))
}

func TestRemoveVolumeIgnoresEmptyID(t *testing.T) {
	svc, mgr := newRemoveTest(dockerConflictErr())

	err := svc.RemoveVolume(context.Background(), "", true, 4, testRetryDelay)

	assert.NoError(t, err)
	assert.Equal(t, 0, mgr.calls)
}

func TestRemoveVolumeStopsWhenContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, mgr := newRemoveTest(dockerConflictErr())
	cancel()

	err := svc.RemoveVolume(ctx, "vol-1", true, 4, time.Hour)

	// One attempt, then the wait is abandoned rather than sat out - without the
	// ctx branch inside gofn this would block for an hour.
	assert.Equal(t, 1, mgr.calls)
	assert.True(t, errors.Is(err, context.Canceled))
}

// The budget is a promise to the caller holding a transaction open, so raising
// any of the constants has to be a deliberate edit rather than a side effect.
func TestRemoveVolumeBudgetStaysSmall(t *testing.T) {
	total := volumeRemovalRetryDelay
	for retry := 1; retry < volumeRemovalRetryMax; retry++ {
		total += volumeRemovalRetryDelay + time.Duration(retry)*volumeRemovalRetryIncr
	}

	assert.Equal(t, 4*time.Second, total)
	assert.Less(t, total, 5*time.Second)
}
