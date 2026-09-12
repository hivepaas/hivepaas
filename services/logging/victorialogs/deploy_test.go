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

// One volume can hold more than logs: the store writes into a directory inside
// it. The volume is still mounted whole, so the mount target does not move.
func TestRuntimeSpecWritesIntoTheDataSubpath(t *testing.T) {
	c := New(&Config{DataVolumeName: "shared", DataSubpath: "logs", Retention: 24 * time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, strings.Join(spec.Args, " "), "-storageDataPath="+DataPath+"/logs")
	assert.Equal(t, DataPath, spec.Mounts[0].Target)
}

func TestRuntimeSpecWithoutSubpathWritesAtTheVolumeRoot(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: 24 * time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, spec.Args, "-storageDataPath="+DataPath)
}

// The value reaches a command line and the store creates what it names, so a
// path leaving the volume would write where nobody asked - inside the
// container, which the next restart throws away.
func TestRuntimeSpecRefusesASubpathThatLeavesTheVolume(t *testing.T) {
	for _, subpath := range []string{"/etc", "..", "../elsewhere", "logs/../..", ".", "./"} {
		_, err := New(&Config{DataVolumeName: "v", DataSubpath: subpath}).RuntimeSpec()
		assert.ErrorIs(t, err, loggingmodel.ErrDataSubpathInvalid, "subpath %q", subpath)
	}

	for _, subpath := range []string{"logs", "hivepaas/logs", "a/../b"} {
		assert.NoError(t, ValidateDataSubpath(subpath), "subpath %q", subpath)
	}
}
