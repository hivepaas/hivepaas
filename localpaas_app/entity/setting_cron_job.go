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
	CurrentCronJobVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCronJob, &cronJobParser{})

type cronJobParser struct {
}

func (s *cronJobParser) New() SettingData {
	return &CronJob{}
}

var (
	parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

type CronJob struct {
	CronType        base.CronJobType         `json:"cronType"`
	Schedule        *CronJobSchedule         `json:"schedule"`
	App             ObjectID                 `json:"app,omitzero"`
	TargetSetting   ObjectID                 `json:"targetSetting,omitzero"`
	Priority        base.TaskPriority        `json:"priority,omitempty"`
	MaxRetry        int                      `json:"maxRetry,omitempty"`
	RetryDelay      timeutil.Duration        `json:"retryDelay,omitempty"`
	Timeout         timeutil.Duration        `json:"timeout,omitempty"`
	ControlDisabled bool                     `json:"controlDisabled,omitempty"`
	Command         *CronJobContainerCommand `json:"command,omitempty"`
	Notification    *BaseEventNotification   `json:"notification,omitempty"`
}

type CronJobSchedule struct {
	CronExpr      string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval      timeutil.Duration `json:"interval,omitempty"`
	InitialTime   time.Time         `json:"initialTime"`
	LastSchedTime time.Time         `json:"lastSchedTime"`
}

func (s *CronJobSchedule) Changed(oldSched *CronJobSchedule) bool {
	return s.CronExpr != oldSched.CronExpr || s.Interval != oldSched.Interval || s.InitialTime != oldSched.InitialTime
}

func (s *CronJobSchedule) OnChange(scheduleChanged bool) {
	if scheduleChanged {
		s.LastSchedTime = time.Time{}
	}
}

func (s *CronJobSchedule) IsValid() error {
	if s.CronExpr != "" {
		if s.Interval > 0 {
			return apperrors.NewValueInvalid()
		}
		_, err := parser.Parse(s.CronExpr)
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

func (s *CronJobSchedule) ParseCronExpr() (cron.Schedule, error) {
	if s.CronExpr == "" {
		return nil, apperrors.NewInactive("Cron expression")
	}
	sched, err := parser.Parse(s.CronExpr)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return sched, nil
}

//nolint:gocognit
func (s *CronJobSchedule) CalcNextRuns(fromTime, toTime time.Time, count int) (res []time.Time, err error) {
	nextRunAt := gofn.Coalesce(s.LastSchedTime, s.InitialTime)
	if toTime.IsZero() && count == 0 {
		return nil, apperrors.NewValueInvalid()
	}

	if s.Interval > 0 {
		interval := s.Interval.ToDuration()
		for {
			if nextRunAt.Before(fromTime) {
				nextRunAt = nextRunAt.Add(interval)
				continue
			}
			if !toTime.IsZero() && nextRunAt.After(toTime) {
				break
			}
			res = append(res, nextRunAt)
			if count > 0 && len(res) >= count {
				break
			}
			nextRunAt = nextRunAt.Add(interval)
		}
		return res, nil
	}

	if s.CronExpr != "" {
		cronSched, err := parser.Parse(s.CronExpr)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for {
			nextRunAt = cronSched.Next(nextRunAt)
			if nextRunAt.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && nextRunAt.After(toTime) {
				break
			}
			res = append(res, nextRunAt)
			if count > 0 && len(res) >= count {
				break
			}
		}
		return res, nil
	}

	return nil, apperrors.NewValueInvalid()
}

type CronJobContainerCommand struct {
	RunInShell string                    `json:"runInShell,omitempty"`
	Command    string                    `json:"command"`
	WorkingDir string                    `json:"workingDir,omitempty"`
	EnvVars    []*EnvVar                 `json:"envVars,omitempty"`
	ArgGroups  []*CronJobCommandArgGroup `json:"argGroups,omitempty"`
}

type CronJobCommandArgGroup struct {
	ExportEnv string               `json:"exportEnv"`
	Separator string               `json:"separator"`
	Args      []*CronJobCommandArg `json:"args,omitempty"`
}

type CronJobCommandArg struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (s *CronJob) GetType() base.SettingType {
	return base.SettingTypeCronJob
}

func (s *CronJob) GetRefObjectIDs() *RefObjectIDs {
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

func (s *CronJob) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCronJobVersion {
		return false, nil
	}
	if setting.Version > CurrentCronJobVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCronJobVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsCronJob() (*CronJob, error) {
	return parseSettingAs[*CronJob](s)
}

func (s *Setting) MustAsCronJob() *CronJob {
	return gofn.Must(s.AsCronJob())
}
