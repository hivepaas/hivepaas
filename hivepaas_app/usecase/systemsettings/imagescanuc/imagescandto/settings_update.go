package imagescandto

import (
	"regexp"
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	repositoryMaxLen = 255

	// The daily scan's interval: no more often than every six hours, which is
	// how often the DB is rebuilt, and no less than weekly.
	minInterval = timeutil.Duration(6 * time.Hour)
	maxInterval = timeutil.Duration(7 * timeutil.Day)

	// maxMemoryLimit and maxCPULimit are sanity bounds, as the registry's are: a
	// typo must not ask swarm for a limit no node could satisfy.
	maxMemoryLimit = 64 * unit.GB
	maxCPULimit    = 256
)

// repositoryPattern is an OCI repository Trivy fetches a DB from, with or
// without its tag: registry.example.com:5000/aquasec/trivy-db:2.
var repositoryPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.\-]*(:[0-9]+)?` + // the registry
	`(/[a-z0-9][a-z0-9._\-]*)+` + // the repository's path
	`(:[A-Za-z0-9_][A-Za-z0-9._\-]*)?$`) // the tag

type UpdateImageScanSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*ImageScanSettingsBaseReq
}

type ImageScanSettingsBaseReq struct {
	Enabled bool `json:"enabled"`
	// DBRepository and JavaDBRepository are OCI repositories of a mirror Trivy
	// fetches its DBs from; empty for Trivy's own.
	DBRepository     string                            `json:"dbRepository"`
	JavaDBRepository string                            `json:"javaDbRepository"`
	Schedule         ScheduleReq                       `json:"schedule"`
	Notification     *basedto.BaseEventNotificationReq `json:"notification"`
	// CPULimit and MemoryLimit are what every scan runs under: the runner's.
	// They live on its service, not in the settings. Zero is the default.
	CPULimit    float64       `json:"cpuLimit"`
	MemoryLimit unit.DataSize `json:"memoryLimit"`
}

// ScheduleReq is the daily scan's: an interval, never a cron.
type ScheduleReq struct {
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func NewUpdateImageScanSettingsReq() *UpdateImageScanSettingsReq {
	return &UpdateImageScanSettingsReq{ImageScanSettingsBaseReq: &ImageScanSettingsBaseReq{}}
}

// ToEntity builds what is stored. A nil request is nil rather than a panic.
func (req *ImageScanSettingsBaseReq) ToEntity() *entity.ImageScanSettings {
	if req == nil {
		return nil
	}
	return &entity.ImageScanSettings{
		Enabled:          req.Enabled,
		DBRepository:     req.DBRepository,
		JavaDBRepository: req.JavaDBRepository,
		Schedule: entity.SchedJobSchedule{
			Interval:    req.Schedule.Interval,
			InitialTime: req.Schedule.InitialTime,
		},
		Notification: req.Notification.ToEntity(),
	}
}

// ScanResources is what the runner is to run under: what was asked, the
// default in place of a zero.
func (req *ImageScanSettingsBaseReq) ScanResources() *systemappservice.Resources {
	res := &systemappservice.Resources{CPULimit: req.CPULimit, MemoryLimit: req.MemoryLimit.Bytes()}
	if res.CPULimit == 0 {
		res.CPULimit = entity.DefaultImageScanCPULimit
	}
	if res.MemoryLimit == 0 {
		res.MemoryLimit = entity.DefaultImageScanMemoryLimit.Bytes()
	}
	return res
}

func (req *UpdateImageScanSettingsReq) ModifyRequest() error {
	if req.ImageScanSettingsBaseReq == nil {
		return nil
	}
	req.DBRepository = strings.TrimSpace(req.DBRepository)
	req.JavaDBRepository = strings.TrimSpace(req.JavaDBRepository)
	if req.Schedule.Interval == 0 {
		req.Schedule.Interval = timeutil.Duration(timeutil.Day)
	}
	if req.Schedule.InitialTime.IsZero() {
		req.Schedule.InitialTime = entity.DefaultImageScanSchedule(timeutil.NowUTC(), timeutil.Location()).InitialTime
	}
	req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Minute)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateImageScanSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	if req.ImageScanSettingsBaseReq == nil {
		return hperrors.NewValidationErrors(vld.Validate(validators...))
	}
	for _, repo := range []struct {
		field string
		value *string
	}{{"dbRepository", &req.DBRepository}, {"javaDbRepository", &req.JavaDBRepository}} {
		validators = append(validators, basedto.ValidateStr(repo.value, false, 1, repositoryMaxLen, repo.field)...)
		if *repo.value != "" {
			validators = append(validators, vld.StrMatch(repo.value, repositoryPattern).
				OnError(vld.SetField(repo.field, nil)))
		}
	}
	validators = append(validators, basedto.ValidateDuration(&req.Schedule.Interval, true, minInterval,
		maxInterval, "schedule.interval")...)
	validators = append(validators, basedto.ValidateTime(&req.Schedule.InitialTime, true,
		time.Now().Add(-timeutil.Dur365Days), time.Time{}, "schedule.initialTime")...)
	validators = append(validators, req.Notification.Validate("notification")...)
	validators = append(validators, basedto.ValidateNumber(&req.MemoryLimit, false,
		entity.MinImageScanMemoryLimit, maxMemoryLimit, "memoryLimit")...)
	validators = append(validators, basedto.ValidateNumber(&req.CPULimit, false, 0, maxCPULimit, "cpuLimit")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateImageScanSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
