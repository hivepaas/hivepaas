package victorialogs

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

func TestRuntimeSpecCarriesRetentionAndStorage(t *testing.T) {
	c := New(&Config{DataVolumeName: "vol-123", Retention: 14 * 24 * time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	args := strings.Join(spec.Args, " ")
	assert.Contains(t, args, "-retentionPeriod=14d")
	assert.Contains(t, args, "-storageDataPath="+DataPath)
	assert.Equal(t, DefaultImage, spec.Image)

	// The data has to outlive the container, so the volume is mounted where
	// -storageDataPath points.
	if len(spec.Mounts) != 1 {
		t.Fatalf("want one mount, got %d", len(spec.Mounts))
	}
	assert.Equal(t, "vol-123", spec.Mounts[0].VolumeName)
	assert.Equal(t, DataPath, spec.Mounts[0].Target)
	assert.False(t, spec.Mounts[0].ReadOnly)
}

// Retention below a day is rejected by VictoriaLogs itself, so rounding down to
// zero days would produce a service that will not start.
func TestRuntimeSpecRoundsRetentionUpToWholeDays(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: 30 * time.Minute})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, strings.Join(spec.Args, " "), "-retentionPeriod=1d")
}

func TestRuntimeSpecOmitsDiskLimitWhenUnset(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.NotContains(t, strings.Join(spec.Args, " "), "maxDiskUsagePercent")
}

func TestRuntimeSpecPassesDiskLimit(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: time.Hour, MaxDiskUsagePercent: 80})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, strings.Join(spec.Args, " "), "-retention.maxDiskUsagePercent=80")
}

func TestRuntimeSpecRequiresAVolume(t *testing.T) {
	c := New(&Config{Retention: time.Hour})

	_, err := c.RuntimeSpec()

	assert.ErrorIs(t, err, loggingmodel.ErrDataVolumeRequired)
}
