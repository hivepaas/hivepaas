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

	// The limits a scan's container runs under. A cold scan of a 773 MB image
	// fitted in 256 MiB and was killed at 128; downloading the DB fitted in 256
	// too. 512 leaves room.
	DefaultImageScanCPULimit    = 1.0
	DefaultImageScanMemoryLimit = 512 * unit.MB
	MinImageScanMemoryLimit     = 256 * unit.MB

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
// scan. It is global, and until Enabled is set nothing is scanned.
type ImageScanSettings struct {
	Enabled bool `json:"enabled,omitempty"`

	// Volume is the cluster volume Trivy keeps its DBs and its layer cache on,
	// required while scanning is on. Its pin decides where scans run: on its
	// node, or, for one every node reaches, on the node running the image.
	Volume ObjectID `json:"volume,omitzero"`

	// CPULimit and MemoryLimit are what each scan's container runs under; zero
	// for the defaults.
	CPULimit    float64       `json:"cpuLimit,omitempty"`
	MemoryLimit unit.DataSize `json:"memoryLimit,omitempty"`

	// DBRepository and JavaDBRepository are where Trivy fetches its DBs from:
	// empty for Trivy's own, an OCI repository of a mirror for an installation
	// without the internet.
	DBRepository     string `json:"dbRepository,omitempty"`
	JavaDBRepository string `json:"javaDbRepository,omitempty"`

	// Schedule and Notification are the daily scan's. A run notifies only when
	// it found something new.
	Schedule     SchedJobSchedule       `json:"schedule"`
	Notification *BaseEventNotification `json:"notification,omitempty"`
}

// ScanCPULimit and ScanMemoryLimit are what a scan's container runs under: the
// settings', or the defaults.
func (s *ImageScanSettings) ScanCPULimit() float64 {
	return gofn.If(s.CPULimit > 0, s.CPULimit, DefaultImageScanCPULimit)
}

func (s *ImageScanSettings) ScanMemoryLimit() unit.DataSize {
	return gofn.If(s.MemoryLimit > 0, s.MemoryLimit, DefaultImageScanMemoryLimit)
}

func (s *ImageScanSettings) GetType() base.SettingType {
	return base.SettingTypeImageScan
}

// GetRefObjectIDs reports the volume while scanning is on, which keeps it from
// being deleted while scans use it, and the notification's targets. Switched
// off, the volume is kept chosen but may go: switching on again asks for one
// that is there.
func (s *ImageScanSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Enabled && s.Volume.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Volume.ID)
	}
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
