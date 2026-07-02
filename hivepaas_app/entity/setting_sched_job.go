package entity

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentSchedJobVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSchedJob, &schedJobParser{})

type schedJobParser struct {
}

func (s *schedJobParser) New() SettingData {
	return &SchedJob{}
}

var (
	cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

type SchedJob struct {
	JobType            base.SchedJobType         `json:"jobType"`
	Schedule           *SchedJobSchedule         `json:"schedule"`
	App                ObjectID                  `json:"app,omitzero"`
	TargetSetting      ObjectID                  `json:"targetSetting,omitzero"`
	Priority           base.TaskPriority         `json:"priority,omitempty"`
	MaxRetry           int                       `json:"maxRetry,omitempty"`
	RetryDelay         timeutil.Duration         `json:"retryDelay,omitempty"`
	RetryDelayIncr     timeutil.Duration         `json:"retryDelayIncr,omitempty"`
	RetryBackoffJitter timeutil.Duration         `json:"retryBackoffJitter,omitempty"`
	RetryDelayMax      timeutil.Duration         `json:"retryDelayMax,omitempty"`
	Timeout            timeutil.Duration         `json:"timeout,omitempty"`
	ControlDisabled    bool                      `json:"controlDisabled,omitempty"`
	Command            *SchedJobContainerCommand `json:"command,omitempty"`
	Notification       *BaseEventNotification    `json:"notification,omitempty"`
}

type SchedJobSchedule struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
	EndTime     time.Time         `json:"endTime,omitzero"`

	InitialTimeAdj  time.Time `json:"initialTimeAdj"`
	LastCronExpr    string    `json:"lastCronExpr,omitempty"`
	LastInitialTime time.Time `json:"lastInitialTime,omitzero"`
}

func (s *SchedJobSchedule) Equal(oldSched *SchedJobSchedule) bool {
	return s.CronExpr == oldSched.CronExpr && s.Interval == oldSched.Interval &&
		s.InitialTime.Equal(oldSched.InitialTime)
}

func (s *SchedJobSchedule) IsValid() error {
	if s.CronExpr != "" {
		if s.Interval > 0 {
			return apperrors.NewArgumentInvalid("Schedule")
		}
		_, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	}
	if s.Interval > 0 {
		return nil
	}
	return apperrors.NewArgumentInvalid("Schedule")
}

func (s *SchedJobSchedule) ParseCronExpr() (cron.Schedule, error) {
	if s.CronExpr == "" {
		return nil, apperrors.NewInactive("Cron expression")
	}
	sched, err := cronParser.Parse(s.CronExpr)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return sched, nil
}

//nolint:gocognit
func (s *SchedJobSchedule) CalcNextRuns(fromTime time.Time, count int) (res []time.Time, err error) {
	if count == 0 {
		return nil, apperrors.NewArgumentInvalid("count")
	}

	nextRunAt := s.InitialTime
	if s.Interval > 0 {
		interval := s.Interval.ToDuration()
		if interval < 0 {
			interval = -interval
		}
		if diff := fromTime.Sub(nextRunAt); diff > interval {
			nextRunAt = nextRunAt.Add((diff / interval) * interval)
		}
		for s.EndTime.IsZero() || nextRunAt.Before(s.EndTime) {
			if nextRunAt.Before(fromTime) {
				nextRunAt = nextRunAt.Add(interval)
				continue
			}
			res = append(res, nextRunAt)
			if len(res) >= count {
				break
			}
			nextRunAt = nextRunAt.Add(interval)
		}
		return res, nil
	}

	if s.CronExpr != "" { //nolint:nestif
		if !s.InitialTimeAdj.IsZero() && s.LastCronExpr == s.CronExpr && s.LastInitialTime.Equal(s.InitialTime) {
			nextRunAt = s.InitialTimeAdj
		}
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, apperrors.New(err)
		}
		for {
			nextRunAt = cronSched.Next(nextRunAt)
			if !s.EndTime.IsZero() && nextRunAt.After(s.EndTime) {
				break
			}
			if nextRunAt.Before(fromTime) {
				continue
			}
			res = append(res, nextRunAt)
			if len(res) >= count {
				break
			}
		}
		return res, nil
	}

	return nil, apperrors.NewArgumentInvalid("Schedule")
}

//nolint:gocognit
func (s *SchedJobSchedule) CalcNextRunsInRange(fromTime, toTime time.Time) (res []time.Time, err error) {
	if toTime.IsZero() {
		return nil, apperrors.NewArgumentInvalid("toTime")
	}
	nextRunAt := s.InitialTime

	if s.Interval > 0 {
		interval := s.Interval.ToDuration()
		if interval < 0 {
			interval = -interval
		}
		if diff := fromTime.Sub(nextRunAt); diff > interval {
			nextRunAt = nextRunAt.Add((diff / interval) * interval)
		}
		for s.EndTime.IsZero() || nextRunAt.Before(s.EndTime) {
			if nextRunAt.Before(fromTime) {
				nextRunAt = nextRunAt.Add(interval)
				continue
			}
			if nextRunAt.After(toTime) {
				break
			}
			res = append(res, nextRunAt)
			nextRunAt = nextRunAt.Add(interval)
		}
		return res, nil
	}

	if s.CronExpr != "" { //nolint:nestif
		if !s.InitialTimeAdj.IsZero() && s.LastCronExpr == s.CronExpr && s.LastInitialTime.Equal(s.InitialTime) {
			nextRunAt = s.InitialTimeAdj
		}
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, apperrors.New(err)
		}
		for {
			nextRunAt = cronSched.Next(nextRunAt)
			if !s.EndTime.IsZero() && nextRunAt.After(s.EndTime) {
				break
			}
			if nextRunAt.Before(fromTime) {
				continue
			}
			if nextRunAt.After(toTime) {
				break
			}
			res = append(res, nextRunAt)
		}
		return res, nil
	}

	return nil, apperrors.NewArgumentInvalid("Schedule")
}

func (s *SchedJobSchedule) AdjustInitialTime(initialTimeAdj time.Time) bool {
	if s.CronExpr == "" { // Only need to adjust initial time on `Cron` mode
		return false
	}
	if !s.InitialTimeAdj.IsZero() && initialTimeAdj.Sub(s.InitialTimeAdj) < timeutil.Dur7Days {
		return false
	}
	s.InitialTimeAdj = initialTimeAdj
	s.LastInitialTime = s.InitialTime
	s.LastCronExpr = s.CronExpr
	return true
}

type SchedJobContainerCommand struct {
	Command     string                     `json:"command"`
	Script      string                     `json:"script,omitempty"`
	WorkingDir  string                     `json:"workingDir,omitempty"`
	EnvVars     []*EnvVar                  `json:"envVars,omitempty"`
	ArgGroups   []*SchedJobCommandArgGroup `json:"argGroups,omitempty"`
	ConsoleSize SchedJobCommandConsoleSize `json:"consoleSize"`
	TTY         bool                       `json:"tty,omitempty"`
	Output      *SchedJobCommandOutput     `json:"output,omitempty"`
}

type SchedJobCommandArgGroup struct {
	Enabled   bool                  `json:"enabled"`
	ExportEnv string                `json:"exportEnv"`
	Separator string                `json:"separator"`
	Args      []*SchedJobCommandArg `json:"args,omitempty"`
}

type SchedJobCommandArg struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SchedJobCommandConsoleSize struct {
	Width  uint `json:"w"`
	Height uint `json:"h"`
}

type SchedJobCommandOutput struct {
	Enabled           bool                       `json:"enabled"`
	SaveFileName      string                     `json:"saveFileName"`
	SavePath          string                     `json:"savePath"`
	Storage           ObjectID                   `json:"storage"`
	FileKind          base.FileKind              `json:"fileKind"`
	CompressionFormat base.FileCompressionFormat `json:"compressionFormat"`
	EncryptionFormat  base.FileEncryptionFormat  `json:"encryptionFormat"`
	EncryptionSecret  EncryptedField             `json:"encryptionSecret"`
}

func (s *SchedJob) GetType() base.SettingType {
	return base.SettingTypeSchedJob
}

func (s *SchedJob) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.App.ID != "" {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.App.ID)
	}
	if s.TargetSetting.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.TargetSetting.ID)
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	if s.Command != nil && s.Command.Output != nil && s.Command.Output.Storage.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Command.Output.Storage.ID)
	}
	return refIDs
}

func (s *SchedJob) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SchedJob) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSchedJobVersion {
		return false, nil
	}
	if setting.Version > CurrentSchedJobVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSchedJobVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSchedJob() (*SchedJob, error) {
	return parseSettingAs[*SchedJob](s)
}

func (s *Setting) MustAsSchedJob() *SchedJob {
	return gofn.Must(s.AsSchedJob())
}
