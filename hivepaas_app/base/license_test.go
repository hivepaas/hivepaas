package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicensePlanDisplayName(t *testing.T) {
	names := map[string]bool{}
	for _, plan := range AllLicensePlans {
		name := plan.DisplayName()
		assert.NotEqual(t, string(plan), name, "plan %q has no display name", plan)
		assert.False(t, names[name], "two plans are called %q", name)
		names[name] = true
	}
	assert.Equal(t, "swarm", LicensePlan("swarm").DisplayName(), "a plan no constant names is shown as it is")
}
