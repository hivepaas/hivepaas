package entity

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	// Sequence is a job-sequence's list of jobs; nil for every other type.
	Sequence *SchedJobSequence `json:"sequence,omitempty"`
	// Triggers are the events that run the job, beside its schedule.
	Triggers []*SchedJobTrigger `json:"triggers,omitempty"`
	// DataBackup is what a data-backup job backs up and where; nil for every other type.
	DataBackup *SchedJobDataBackup `json:"dataBackup,omitempty"`
	// FunctionInvoke is the request a function-invoke job calls its function
	// with; nil for every other type.
	FunctionInvoke *SchedJobFunctionInvoke `json:"functionInvoke,omitempty"`
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
	if s == nil || oldSched == nil {
		return s == nil && oldSched == nil
	}
	return s.CronExpr == oldSched.CronExpr && s.Interval == oldSched.Interval &&
		s.InitialTime.Equal(oldSched.InitialTime) && s.EndTime.Equal(oldSched.EndTime)
}

func (s *SchedJobSchedule) IsValid() error {
	if s.CronExpr != "" {
		if s.Interval > 0 {
			return hperrors.NewArgumentInvalid("Schedule")
		}
		_, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	}
	if s.Interval > 0 {
		return nil
	}
	return hperrors.NewArgumentInvalid("Schedule")
}

func (s *SchedJobSchedule) GetLastSchedTime() time.Time {
	if !s.LastSchedTime.IsZero() && s.LastCronExpr == s.CronExpr && s.LastInterval == s.Interval &&
		s.LastInitialTime.Equal(s.LastSchedTime) {
		return s.LastSchedTime
	}
	return s.InitialTime
}

func (s *SchedJobSchedule) SetLastSchedTime(lastSchedTime time.Time) bool {
	if s == nil {
		return false
	}
	if s.LastSchedTime.Equal(lastSchedTime) {
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
		return nil, hperrors.NewInactive("Cron expression")
	}
	sched, err := cronParser.Parse(s.CronExpr)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return sched, nil
}

// cronLocation is where a cron expression's hours are read: in the zone of the
// initial time when it was given with an offset of its own - an API client's
// 2026-09-27T00:00:00+07:00 - and otherwise, for one in UTC as the dashboard
// gives it, in the installation's timezone, daylight saving and all.
func (s *SchedJobSchedule) cronLocation() *time.Location {
	if _, offset := s.InitialTime.Zone(); offset != 0 {
		return s.InitialTime.Location()
	}
	return timeutil.Location()
}

// CalcNextRuns is the next count runs from fromTime. A job without a schedule
// has none: it runs only by hand or as a step of a job sequence.
func (s *SchedJobSchedule) CalcNextRuns(fromTime time.Time, count int) (res []time.Time, err error) {
	if s == nil {
		return nil, nil
	}
	return s.calcNextRuns(fromTime, count)
}

func (s *SchedJobSchedule) calcNextRuns(fromTime time.Time, count int) (res []time.Time, err error) {
	if count == 0 {
		return nil, hperrors.NewArgumentInvalid("count")
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

	if s.CronExpr != "" {
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		// Next reads the expression's hours in the location of the time it is
		// given: see cronLocation.
		nextRunAt = nextRunAt.In(s.cronLocation())
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

	return nil, hperrors.NewArgumentInvalid("Schedule")
}

// CalcNextRunsInRange is the runs between fromTime and toTime; none for a job
// without a schedule.
func (s *SchedJobSchedule) CalcNextRunsInRange(fromTime, toTime time.Time) (res []time.Time, err error) {
	if s == nil {
		return nil, nil
	}
	return s.calcNextRunsInRange(fromTime, toTime)
}

func (s *SchedJobSchedule) calcNextRunsInRange(fromTime, toTime time.Time) (res []time.Time, err error) {
	if toTime.IsZero() {
		return nil, hperrors.NewArgumentInvalid("toTime")
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

	if s.CronExpr != "" {
		cronSched, err := cronParser.Parse(s.CronExpr)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		// Next reads the expression's hours in the location of the time it is
		// given: see cronLocation.
		nextRunAt = nextRunAt.In(s.cronLocation())
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

	return nil, hperrors.NewArgumentInvalid("Schedule")
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
	if s.Command != nil {
		if s.Command.Script.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Command.Script.ID)
		}
	}
	if s.CommandOutput != nil && s.CommandOutput.SaveToFile != nil {
		if s.CommandOutput.SaveToFile.Storage.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.CommandOutput.SaveToFile.Storage.ID)
		}
	}
	if s.CommandOutput != nil && s.CommandOutput.PipeToApp != nil {
		if s.CommandOutput.PipeToApp.TargetApp.ID != "" {
			refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.CommandOutput.PipeToApp.TargetApp.ID)
		}
		if s.CommandOutput.PipeToApp.Command != nil && s.CommandOutput.PipeToApp.Command.Script.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.CommandOutput.PipeToApp.Command.Script.ID)
		}
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	// A sequence's members are references like any other: they are what keeps a
	// job a sequence runs from being deleted under it.
	refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Sequence.MemberIDs()...)
	refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.TriggerAppIDs()...)
	refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.DataBackup.refSettingIDs()...)
	return refIDs
}

func (s *SchedJob) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsSchedJob() (*SchedJob, error) {
	return parseSettingAs[*SchedJob](s)
}

func (s *Setting) MustAsSchedJob() *SchedJob {
	return gofn.Must(s.AsSchedJob())
}
