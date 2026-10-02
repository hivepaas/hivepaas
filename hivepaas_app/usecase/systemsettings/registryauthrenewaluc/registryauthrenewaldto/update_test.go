package registryauthrenewaldto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// The interval is from 1 to 10 hours: no 12-hour token is fresh enough for a
// longer one.
func TestUpdateRegistryAuthRenewalIntervalBounds(t *testing.T) {
	for interval, valid := range map[time.Duration]bool{
		0:                false,
		59 * time.Minute: false,
		time.Hour:        true,
		6 * time.Hour:    true,
		10 * time.Hour:   true,
		11 * time.Hour:   false,
	} {
		req := &RegistryAuthRenewalBaseReq{
			Status:   base.SettingStatusActive,
			Schedule: ScheduleReq{Interval: timeutil.Duration(interval), InitialTime: time.Now()},
		}
		assert.Equal(t, valid, len(vld.Validate(req.validate("")...)) == 0, interval.String())
	}
}
