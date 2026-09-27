package taskschedjobexec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func TestDescribeSchedule(t *testing.T) {
	assert.Equal(t, "no schedule (run by hand or by a job sequence)", describeSchedule(nil))
	assert.Equal(t, "every 1h", describeSchedule(&entity.SchedJobSchedule{Interval: timeutil.Duration(3600e9)}))
	assert.Equal(t, "cron expression 0 * * * *", describeSchedule(&entity.SchedJobSchedule{CronExpr: "0 * * * *"}))
}
