package imagescandto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func updateReq(base ImageScanSettingsBaseReq) *UpdateImageScanSettingsReq {
	req := NewUpdateImageScanSettingsReq()
	*req.ImageScanSettingsBaseReq = base
	_ = req.ModifyRequest()
	return req
}

// A save with nothing but the switch gets the daily schedule and the default
// limits; a mirror is an OCI repository.
func TestUpdateImageScanSettingsDefaults(t *testing.T) {
	req := updateReq(ImageScanSettingsBaseReq{Enabled: true})
	assert.Empty(t, req.Validate())
	assert.Equal(t, timeutil.Duration(timeutil.Day), req.Schedule.Interval)
	assert.Equal(t, entity.ImageScanDailyHour, req.Schedule.InitialTime.In(timeutil.Location()).Hour())
	res := req.ScanResources()
	assert.Equal(t, entity.DefaultImageScanCPULimit, res.CPULimit)
	assert.Equal(t, entity.DefaultImageScanMemoryLimit.Bytes(), res.MemoryLimit)

	for _, repo := range []string{"mirror.local:5000/aquasec/trivy-db:2", "ghcr.io/aquasecurity/trivy-db",
		"registry.example.com/a/b/c"} {
		assert.Empty(t, updateReq(ImageScanSettingsBaseReq{DBRepository: repo}).Validate(), repo)
	}
	for _, repo := range []string{"trivy-db", "mirror.local/x; rm -rf /", "mirror.local/x y", "https://m/x"} {
		assert.NotEmpty(t, updateReq(ImageScanSettingsBaseReq{JavaDBRepository: repo}).Validate(), repo)
	}
}

// The schedule is six hours to a week; a scan's memory leaves Trivy room.
func TestUpdateImageScanSettingsBounds(t *testing.T) {
	assert.NotEmpty(t, updateReq(ImageScanSettingsBaseReq{
		Schedule: ScheduleReq{Interval: timeutil.Duration(time.Hour)}}).Validate())
	assert.NotEmpty(t, updateReq(ImageScanSettingsBaseReq{
		Schedule: ScheduleReq{Interval: timeutil.Duration(8 * timeutil.Day)}}).Validate())
	assert.NotEmpty(t, updateReq(ImageScanSettingsBaseReq{MemoryLimit: 128 * unit.MB}).Validate())
	assert.Empty(t, updateReq(ImageScanSettingsBaseReq{MemoryLimit: 256 * unit.MB, CPULimit: 0.5}).Validate())
}
