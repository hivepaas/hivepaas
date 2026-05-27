package entity

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
	JobType         base.SchedJobType         `json:"jobType"`
	Schedule        *SchedJobSchedule         `json:"schedule"`
	App             ObjectID                  `json:"app,omitzero"`
	TargetSetting   ObjectID                  `json:"targetSetting,omitzero"`
	Priority        base.TaskPriority         `json:"priority,omitempty"`
	MaxRetry        int                       `json:"maxRetry,omitempty"`
	RetryDelay      timeutil.Duration         `json:"retryDelay,omitempty"`
	Timeout         timeutil.Duration         `json:"timeout,omitempty"`
	ControlDisabled bool                      `json:"controlDisabled,omitempty"`
	Command         *SchedJobContainerCommand `json:"command,omitempty"`
	Notification    *BaseEventNotification    `json:"notification,omitempty"`
}

type SchedJobSchedule struct {
	CronExpr      string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval      timeutil.Duration `json:"interval,omitempty"`
	InitialTime   time.Time         `json:"initialTime"`
	LastSchedTime time.Time         `json:"lastSchedTime"`
}

func (s *SchedJobSchedule) Changed(oldSched *SchedJobSchedule) bool {
	return s.CronExpr != oldSched.CronExpr || s.Interval != oldSched.Interval || s.InitialTime != oldSched.InitialTime
}

func (s *SchedJobSchedule) OnChange(scheduleChanged bool) {
	if scheduleChanged {
		s.LastSchedTime = time.Time{}
	}
}

func (s *SchedJobSchedule) IsValid() error {
	if s.CronExpr != "" {
		if s.Interval > 0 {
			return apperrors.NewValueInvalid()
		}
		_, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}
	if s.Interval > 0 {
		return nil
	}
	return apperrors.NewValueInvalid()
}

func (s *SchedJobSchedule) ParseCronExpr() (cron.Schedule, error) {
	if s.CronExpr == "" {
		return nil, apperrors.NewInactive("Cron expression")
	}
	sched, err := cronParser.Parse(s.CronExpr)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return sched, nil
}

func (s *SchedJobSchedule) CalcNextRuns(fromTime time.Time, count int) (res []time.Time, err error) {
	nextRunAt := gofn.Coalesce(s.LastSchedTime, s.InitialTime)
	if count == 0 {
		return nil, apperrors.NewValueInvalid()
	}

	if s.Interval > 0 {
		interval := s.Interval.ToDuration()
		for {
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

	if s.CronExpr != "" {
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for {
			nextRunAt = cronSched.Next(nextRunAt)
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

	return nil, apperrors.NewValueInvalid()
}

func (s *SchedJobSchedule) CalcNextRunsInRange(fromTime, toTime time.Time) (res []time.Time, err error) {
	nextRunAt := gofn.Coalesce(s.LastSchedTime, s.InitialTime)
	if toTime.IsZero() {
		return nil, apperrors.NewValueInvalid()
	}

	if s.Interval > 0 {
		interval := s.Interval.ToDuration()
		for {
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

	if s.CronExpr != "" {
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for {
			nextRunAt = cronSched.Next(nextRunAt)
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

	return nil, apperrors.NewValueInvalid()
}

type SchedJobContainerCommand struct {
	RunInShell string                     `json:"runInShell,omitempty"`
	Command    string                     `json:"command"`
	WorkingDir string                     `json:"workingDir,omitempty"`
	EnvVars    []*EnvVar                  `json:"envVars,omitempty"`
	ArgGroups  []*SchedJobCommandArgGroup `json:"argGroups,omitempty"`
}

type SchedJobCommandArgGroup struct {
	ExportEnv string                `json:"exportEnv"`
	Separator string                `json:"separator"`
	Args      []*SchedJobCommandArg `json:"args,omitempty"`
}

type SchedJobCommandArg struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
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
	return refIDs
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
