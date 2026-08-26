package dockerhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutRestrictedLabels(t *testing.T) {
	input := map[string]string{
		"com.docker.stack.namespace": "system-stack",
		"hivepaas.service.name":      "app",
		"traefik.enable":             "true",
		"custom.label":               "value1",
		"my-org/env":                 "production",
	}

	expected := map[string]string{
		"custom.label": "value1",
		"my-org/env":   "production",
	}

	got := FilterOutRestrictedLabels(input)
	assert.Equal(t, expected, got)
}

func TestValidateUserLabels(t *testing.T) {
	t.Run("no violations", func(t *testing.T) {
		input := map[string]string{
			"custom.label": "value1",
			"my-org/env":   "production",
		}
		got := ValidateUserLabels(input, false)
		assert.Empty(t, got)
	})

	t.Run("stop at first violation", func(t *testing.T) {
		input := map[string]string{
			"com.docker.stack.namespace": "system-stack",
			"hivepaas.service.name":      "app",
			"custom.label":               "value1",
		}
		got := ValidateUserLabels(input, true)
		assert.Len(t, got, 1)
		assert.Contains(t, []string{"com.docker.stack.namespace", "hivepaas.service.name"}, got[0])
	})

	t.Run("collect all violations", func(t *testing.T) {
		input := map[string]string{
			"com.docker.stack.namespace": "system-stack",
			"hivepaas.service.name":      "app",
			"traefik.enable":             "true",
			"custom.label":               "value1",
		}
		expected := []string{
			"com.docker.stack.namespace",
			"hivepaas.service.name",
			"traefik.enable",
		}
		got := ValidateUserLabels(input, false)
		assert.ElementsMatch(t, expected, got)
	})
}

func TestApplyUserLabels(t *testing.T) {
	currLabels := map[string]string{
		"com.docker.stack.namespace": "system-stack",
		"hivepaas.service.name":      "old-app",
		"custom.old":                 "should-be-removed",
	}

	userLabels := map[string]string{
		"hivepaas.service.name": "malicious-attempt", // restricted, should be ignored
		"traefik.enable":        "false",             // restricted, should be ignored
		"custom.new":            "value-new",         // allowed, should be applied
		"custom.updated":        "value-updated",     // allowed, should be applied
	}

	expected := map[string]string{
		"com.docker.stack.namespace": "system-stack",
		"hivepaas.service.name":      "old-app",
		"custom.new":                 "value-new",
		"custom.updated":             "value-updated",
	}

	got := ApplyUserLabels(currLabels, userLabels)
	assert.Equal(t, expected, got)
}
