package sysupdateserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// runMajorUpgrade drives one step through updateServiceImage, which is where the
// release's say over major versions is applied.
func runMajorUpgrade(t *testing.T, release *base.ReleaseInfo,
	component, currentImage, targetImage string,
) (*fakeDocker, error) {
	t.Helper()

	f := &fakeDocker{}
	s := &service{dockerManager: f}
	err := s.updateServiceImage(context.Background(), loggingUpdateData(t, release), serviceImageUpdate{
		What:        component,
		Component:   component,
		TargetImage: targetImage,
		Fetch: func(_ context.Context) (*swarm.Service, error) {
			return swarmServiceOn(component, currentImage, 1), nil
		},
	})
	return f, err
}

// Postgres will not start on a data directory written by a different major, so
// the release lists it. Without the refusal the update swaps the image, postgres
// declines to come up, and the discovery happens with the system already down.
func TestMajorUpgradeIsRefusedForAComponentTheReleaseLists(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	f, err := runMajorUpgrade(t, release, base.HivepaasDbKey,
		"postgres:18.3-alpine", "postgres:19.0-alpine")

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
	assert.Empty(t, f.updated, "nothing is applied when the step is refused")
}

func TestMajorUpgradeAllowsAMinorChangeOnTheSameComponent(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	_, err := runMajorUpgrade(t, release, base.HivepaasDbKey,
		"postgres:18.3-alpine", "postgres:18.6-alpine")

	assert.NoError(t, err)
}

// The default is that nothing is blocked. Crossing a major only matters where a
// component owns durable state it might convert, and the release is what decides
// which those are - VictoriaLogs is deliberately not one of them.
func TestMajorUpgradeIsAllowedForEveryComponentTheReleaseDoesNotList(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	f, err := runMajorUpgrade(t, release, base.HivepaasVictoriaLogsKey,
		"victoriametrics/victoria-logs:v1.52.0", "victoriametrics/victoria-logs:v2.0.0")

	assert.NoError(t, err)
	assert.Equal(t, "victoriametrics/victoria-logs:v2.0.0", f.updated[base.HivepaasVictoriaLogsKey])
}

// A running service reports its image with the digest it was pinned to, which
// must not hide the major version in the tag.
func TestMajorUpgradeReadsTheMajorThroughADigest(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	_, err := runMajorUpgrade(t, release, base.HivepaasDbKey,
		"postgres:18.3-alpine@sha256:aaaa", "postgres:19.0-alpine")

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

// Deliberate: refusing every release that names a tag this cannot parse would be
// the more expensive mistake. The check exists for one specific event, not as a
// gate on all image changes.
func TestMajorUpgradeAllowsWhatItCannotRead(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	_, err := runMajorUpgrade(t, release, base.HivepaasDbKey, "postgres:latest", "postgres:19.0-alpine")
	assert.NoError(t, err)

	_, err = runMajorUpgrade(t, release, base.HivepaasDbKey, "postgres:18.3-alpine", "postgres:latest")
	assert.NoError(t, err)
}

// A step with no component key is outside the release's reach entirely.
func TestMajorUpgradeIgnoresAStepWithNoComponent(t *testing.T) {
	release := &base.ReleaseInfo{BlockMajorUpgrade: []string{base.HivepaasDbKey}}

	f := &fakeDocker{}
	s := &service{dockerManager: f}
	err := s.updateServiceImage(context.Background(), loggingUpdateData(t, release), serviceImageUpdate{
		What:        "something",
		TargetImage: "postgres:19.0-alpine",
		Fetch: func(_ context.Context) (*swarm.Service, error) {
			return swarmServiceOn("something", "postgres:18.3-alpine", 1), nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "postgres:19.0-alpine", f.updated["something"])
}
