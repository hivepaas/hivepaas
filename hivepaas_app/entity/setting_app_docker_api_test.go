package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func TestAppDockerAPISettingsReadWhatWasStored(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeAppDockerAPI, Data: `{"images":["autobase/automation:2.11.0"],` +
		`"sharedDirs":["/var/lib/autobase/ansible"],"networks":["env"],"allow":["exec"],` +
		`"limits":{"containers":3,"memory":"2gb","cpus":2}}`}

	got, err := setting.AsAppDockerAPISettings()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, &AppDockerAPISettings{
		Images:     []string{"autobase/automation:2.11.0"},
		SharedDirs: []string{"/var/lib/autobase/ansible"},
		Networks:   []string{DockerAPINetworkEnv},
		Allow:      []string{"exec"},
		Limits:     AppDockerAPILimits{Containers: 3, Memory: 2 * unit.GB, CPUs: 2},
	}, got)
}
