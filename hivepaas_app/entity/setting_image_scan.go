package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentImageScanSettingsVersion = 1

	// The limits a scan runs under: those of the runner, the container on each
	// node every scan of that node runs in. A cold scan of a 773 MB image fitted
	// in 256 MiB and was killed at 128; 512 leaves room.
	DefaultImageScanCPULimit    = 1.0
	DefaultImageScanMemoryLimit = 512 * unit.MB
	MinImageScanMemoryLimit     = 256 * unit.MB

	// The server's limits. It holds the DB memory-mapped, about 80 MiB, and
	// needs about 512 MiB while it swaps in a new one.
	DefaultImageScanServerCPULimit    = 1.0
	DefaultImageScanServerMemoryLimit = 1 * unit.GB

	// ImageScanDailyHour is the time of day the daily scan runs at, in the
	// installation's timezone: after the other daily jobs, in the quiet hours.
	ImageScanDailyHour = 2
)

// DefaultImageScanSchedule is the daily scan's schedule when nothing says
// otherwise: every 24 hours from ImageScanDailyHour in loc, on the day now is
// there.
func DefaultImageScanSchedule(now time.Time, loc *time.Location) SchedJobSchedule {
	local := now.In(loc)
	return SchedJobSchedule{
		Interval: timeutil.Duration(timeutil.Day),
		InitialTime: time.Date(local.Year(), local.Month(), local.Day(), ImageScanDailyHour, 0, 0, 0,
			loc).UTC(),
	}
}

var _ = registerSettingParser(base.SettingTypeImageScan, &imageScanSettingsParser{})

type imageScanSettingsParser struct {
}

func (s *imageScanSettingsParser) New() SettingData {
	return &ImageScanSettings{}
}

// ImageScanSettings is image scanning's configuration: the addon, and the daily
// scan. It is global, and until Enabled is set nothing is provisioned and
// nothing is scanned.
type ImageScanSettings struct {
	Enabled bool `json:"enabled,omitempty"`

	// DBRepository and JavaDBRepository are where Trivy fetches its DBs from:
	// empty for Trivy's own, an OCI repository of a mirror for an installation
	// without the internet.
	DBRepository     string `json:"dbRepository,omitempty"`
	JavaDBRepository string `json:"javaDbRepository,omitempty"`

	// Schedule and Notification are the daily scan's. A run notifies only when
	// it found something new.
	Schedule     SchedJobSchedule       `json:"schedule"`
	Notification *BaseEventNotification `json:"notification,omitempty"`

	// ServerAppID and RunnerAppID are the apps the addon made, written back by
	// it for the dashboard to link to. They are not the client's.
	ServerAppID string `json:"serverAppId,omitempty"`
	RunnerAppID string `json:"runnerAppId,omitempty"`
}

func (s *ImageScanSettings) GetType() base.SettingType {
	return base.SettingTypeImageScan
}

func (s *ImageScanSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *ImageScanSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsImageScanSettings() (*ImageScanSettings, error) {
	return parseSettingAs[*ImageScanSettings](s)
}

func (s *Setting) MustAsImageScanSettings() *ImageScanSettings {
	return gofn.Must(s.AsImageScanSettings())
}

func (s *ImageScanSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentImageScanSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentImageScanSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	setting.Version = CurrentImageScanSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
