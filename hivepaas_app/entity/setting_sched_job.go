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
	JobType            base.SchedJobType      `json:"jobType"`
	Schedule           *SchedJobSchedule      `json:"schedule"`
	App                ObjectID               `json:"app,omitzero"`
	TargetSetting      ObjectID               `json:"targetSetting,omitzero"`
	Priority           base.TaskPriority      `json:"priority,omitempty"`
	MaxRetry           int                    `json:"maxRetry,omitempty"`
	RetryDelay         timeutil.Duration      `json:"retryDelay,omitempty"`
	RetryDelayIncr     timeutil.Duration      `json:"retryDelayIncr,omitempty"`
	RetryBackoffJitter timeutil.Duration      `json:"retryBackoffJitter,omitempty"`
	RetryDelayMax      timeutil.Duration      `json:"retryDelayMax,omitempty"`
	Timeout            timeutil.Duration      `json:"timeout,omitempty"`
	ControlDisabled    bool                   `json:"controlDisabled,omitempty"`
	Command            *CommandTemplate       `json:"command,omitempty"`
	CommandOutput      *SchedJobCommandOutput `json:"commandOutput,omitempty"`
	Notification       *BaseEventNotification `json:"notification,omitempty"`
}

type SchedJobSchedule struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
	EndTime     time.Time         `json:"endTime,omitzero"`

	LastSchedTime   time.Time         `json:"lastSchedTime"`
	LastCronExpr    string            `json:"lastCronExpr,omitempty"`
	LastInterval    timeutil.Duration `json:"lastInterval,omitempty"`
	LastInitialTime time.Time         `json:"lastInitialTime,omitzero"`
}

func (s *SchedJobSchedule) Equal(oldSched *SchedJobSchedule) bool {
	return s.CronExpr == oldSched.CronExpr && s.Interval == oldSched.Interval &&
		s.InitialTime.Equal(oldSched.InitialTime) && s.EndTime.Equal(oldSched.EndTime)
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

func (s *SchedJobSchedule) GetLastSchedTime() time.Time {
	if !s.LastSchedTime.IsZero() && s.LastCronExpr == s.CronExpr && s.LastInterval == s.Interval &&
		s.LastInitialTime.Equal(s.LastSchedTime) {
		return s.LastSchedTime
	}
	return s.InitialTime
}

func (s *SchedJobSchedule) SetLastSchedTime(lastSchedTime time.Time) bool {
	if !s.LastSchedTime.IsZero() && lastSchedTime.Sub(s.LastSchedTime) < timeutil.Day {
		return false
	}
	s.LastSchedTime = lastSchedTime
	s.LastCronExpr = s.CronExpr
	s.LastInterval = s.Interval
	s.LastInitialTime = s.InitialTime
	return true
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

func (s *SchedJobSchedule) CalcNextRuns(fromTime time.Time, count int) (res []time.Time, err error) {
	if count == 0 {
		return nil, apperrors.NewArgumentInvalid("count")
	}

	nextRunAt := s.GetLastSchedTime()
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

func (s *SchedJobSchedule) CalcNextRunsInRange(fromTime, toTime time.Time) (res []time.Time, err error) {
	if toTime.IsZero() {
		return nil, apperrors.NewArgumentInvalid("toTime")
	}
	nextRunAt := s.GetLastSchedTime()

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

type SchedJobCommandOutput struct {
	Enabled    bool                             `json:"enabled"`
	SaveToFile *SchedJobCommandOutputSaveToFile `json:"saveToFile,omitempty"`
	PipeToApp  *SchedJobCommandOutputPipeToApp  `json:"pipeToApp,omitempty"`
}

type SchedJobCommandOutputSaveToFile struct {
	FileName          string                     `json:"fileName"`
	FilePath          string                     `json:"filePath"`
	FileKind          base.FileKind              `json:"fileKind"`
	Storage           ObjectID                   `json:"storage"`
	CompressionFormat base.FileCompressionFormat `json:"compressionFormat"`
	EncryptionFormat  base.FileEncryptionFormat  `json:"encryptionFormat"`
	EncryptionSecret  EncryptedField             `json:"encryptionSecret"`
}

type SchedJobCommandOutputPipeToApp struct {
	TargetApp ObjectID         `json:"targetApp"`
	Command   *CommandTemplate `json:"command"`
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
	if s.CommandOutput != nil {
		if s.CommandOutput.SaveToFile != nil && s.CommandOutput.SaveToFile.Storage.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.CommandOutput.SaveToFile.Storage.ID)
		}
		if s.CommandOutput.PipeToApp != nil && s.CommandOutput.PipeToApp.TargetApp.ID != "" {
			refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.CommandOutput.PipeToApp.TargetApp.ID)
		}
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
