package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

func TestNewBackendBuildsVictoriaLogs(t *testing.T) {
	b, err := NewBackend(BackendTypeVictoriaLogs, &BackendConfig{VictoriaLogs: &victorialogs.Config{}})

	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestNewBackendRejectsAnUnknownType(t *testing.T) {
	_, err := NewBackend("loki", &BackendConfig{VictoriaLogs: &victorialogs.Config{}})

	assert.ErrorIs(t, err, ErrBackendUnsupported)
}

func TestNewDeployerReturnsSomethingThatCanDescribeItself(t *testing.T) {
	d, err := NewDeployer(BackendTypeVictoriaLogs, &BackendConfig{
		VictoriaLogs: &victorialogs.Config{DataVolumeName: "v"},
	})
	if err != nil {
		t.Fatalf("NewDeployer: %v", err)
	}

	spec, err := d.RuntimeSpec()
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Image)
}

func TestNewCollectorBuildsVlagent(t *testing.T) {
	c, err := NewCollector(CollectorTypeVlagent, &CollectorConfig{Vlagent: &vlagent.Config{}})

	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewCollectorRejectsAnUnknownType(t *testing.T) {
	_, err := NewCollector("fluent-bit", &CollectorConfig{Vlagent: &vlagent.Config{}})

	assert.ErrorIs(t, err, ErrCollectorUnsupported)
}
