package licenseservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// A function's clone is a function only with its deployment settings; without
// them it has no kind, and is an app.
func TestCloneLimit(t *testing.T) {
	assert.Equal(t, base.LicenseLimitFunctions, CloneLimit(true, true))
	assert.Equal(t, base.LicenseLimitApps, CloneLimit(true, false))
	assert.Equal(t, base.LicenseLimitApps, CloneLimit(false, true))
	assert.Equal(t, base.LicenseLimitApps, CloneLimit(false, false))
}

func TestAddingIsOneKind(t *testing.T) {
	assert.Equal(t, &CheckReq{Adding: map[base.LicenseLimit]int{base.LicenseLimitUsers: 2}},
		Adding(base.LicenseLimitUsers, 2))
}
