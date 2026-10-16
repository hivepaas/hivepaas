package licenseserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Usage counts every limit, for the dashboard to show against the plan's, and
// locks nothing: it creates nothing.
func TestUsageCountsEveryLimit(t *testing.T) {
	h := newHarness(t)
	h.docker.nodes = 3
	h.usage.oauth = 1
	h.usage.counts = map[base.LicenseLimit]int{base.LicenseLimitUsers: 4, base.LicenseLimitApps: 25,
		base.LicenseLimitFunctions: 2, base.LicenseLimitPreviews: 1}

	usage, err := h.Usage(context.Background(), (*database.DB)(nil))

	assert.NoError(t, err)
	assert.Equal(t, map[base.LicenseLimit]int{
		base.LicenseLimitNodes:           3,
		base.LicenseLimitUsers:           4,
		base.LicenseLimitApps:            25,
		base.LicenseLimitFunctions:       2,
		base.LicenseLimitPreviews:        1,
		base.LicenseLimitSignInProviders: 1,
	}, usage.Used)
	assert.Len(t, usage.Used, len(base.AllLicenseLimits))
	assert.Equal(t, []string{"count users", "count apps", "count functions", "count previews",
		"count oauth [google microsoft-online openid-connect]"}, h.usage.asked)
}
