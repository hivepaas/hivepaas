package imagescandto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const nextRunsCount = 5

type GetImageScanSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetImageScanSettingsReq() *GetImageScanSettingsReq {
	return &GetImageScanSettingsReq{}
}

func (req *GetImageScanSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetImageScanSettingsResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *ImageScanSettingsResp `json:"data"`
}

type ImageScanSettingsResp struct {
	*settings.BaseSettingResp

	Enabled          bool                               `json:"enabled"`
	DBRepository     string                             `json:"dbRepository"`
	JavaDBRepository string                             `json:"javaDbRepository"`
	Schedule         *ScheduleResp                      `json:"schedule"`
	Notification     *basedto.BaseEventNotificationResp `json:"notification"`
	// CPULimit and MemoryLimit are read off the runner's service, where they
	// live. With no runner they are what a new one would get.
	CPULimit    float64       `json:"cpuLimit"`
	MemoryLimit unit.DataSize `json:"memoryLimit"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
	// ScannerStatus is what runs, as opposed to what was configured. It is not
	// called Status: BaseSettingResp has one.
	ScannerStatus *ScannerStatusResp `json:"scannerStatus"`
}

// ScheduleResp is the daily scan's schedule: an interval, never a cron.
type ScheduleResp struct {
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

// ScannerStatusResp is what runs: the server, the runner on the nodes, and
// the DB of the last scan.
type ScannerStatusResp struct {
	ServerAppID   string `json:"serverAppId,omitempty"`
	RunnerAppID   string `json:"runnerAppId,omitempty"`
	ServerRunning bool   `json:"serverRunning"`
	// RunnerNodes is on how many of Nodes the runner runs.
	RunnerNodes  int        `json:"runnerNodes"`
	Nodes        int        `json:"nodes"`
	Image        string     `json:"image"`
	TrivyVersion string     `json:"trivyVersion,omitempty"`
	DBUpdatedAt  *time.Time `json:"dbUpdatedAt,omitempty"`
	LastScanAt   *time.Time `json:"lastScanAt,omitempty"`
}

func TransformImageScanSettings(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
	status *imagescanservice.Status,
) (resp *ImageScanSettingsResp, err error) {
	resp = &ImageScanSettingsResp{}
	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// An installation that has not saved the settings yet answers the defaults.
	cfg := &entity.ImageScanSettings{
		Schedule: entity.DefaultImageScanSchedule(timeutil.NowUTC(), timeutil.Location()),
		Notification: &entity.BaseEventNotification{
			SuccessUseDefault: true,
			FailureUseDefault: true,
		},
	}
	if setting.Data != "" {
		if cfg, err = setting.AsImageScanSettings(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	resp.Enabled = cfg.Enabled
	resp.DBRepository, resp.JavaDBRepository = cfg.DBRepository, cfg.JavaDBRepository
	resp.Schedule = &ScheduleResp{Interval: cfg.Schedule.Interval, InitialTime: cfg.Schedule.InitialTime}
	resp.Notification = basedto.TransformBaseEventNotification(cfg.Notification, refObjects)
	if cfg.Enabled {
		resp.NextRuns, _ = cfg.Schedule.CalcNextRuns(time.Now(), nextRunsCount) //nolint:errcheck
	}

	if status != nil {
		resp.CPULimit = status.ScanResources.CPULimit
		resp.MemoryLimit = unit.DataSize(status.ScanResources.MemoryLimit)
		resp.ScannerStatus = &ScannerStatusResp{
			ServerAppID:   status.ServerAppID,
			RunnerAppID:   status.RunnerAppID,
			ServerRunning: status.ServerRunning,
			RunnerNodes:   status.RunnerNodes,
			Nodes:         status.Nodes,
			Image:         status.Image,
			TrivyVersion:  status.TrivyVersion,
			DBUpdatedAt:   timeOrNil(status.DBUpdatedAt),
			LastScanAt:    timeOrNil(status.LastScanAt),
		}
	}
	return resp, nil
}

func timeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
