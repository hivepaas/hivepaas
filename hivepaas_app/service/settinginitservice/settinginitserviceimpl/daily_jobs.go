package settinginitserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// dailyJob is a system job HivePaaS runs once a day: by default every 24 hours
// from a time of day in the installation's timezone, in the quiet hours after
// midnight, each apart from the others.
type dailyJob struct {
	settingType  base.SettingType
	hour, minute int
	// schedule is the job's schedule within its setting's data, and the data
	// to write back once it changed.
	schedule func(setting *entity.Setting) (*entity.SchedJobSchedule, entity.SettingData, error)
}

var dailyJobs = []dailyJob{
	{base.SettingTypeSystemCleanup, 0, 0, func(s *entity.Setting) (*entity.SchedJobSchedule, entity.SettingData, error) {
		data, err := s.AsSystemCleanup()
		return scheduleOf(data, err, func() *entity.SchedJobSchedule { return &data.Schedule })
	}},
	{base.SettingTypeSystemBackup, 0, 30, func(s *entity.Setting) (*entity.SchedJobSchedule, entity.SettingData, error) {
		data, err := s.AsSystemBackup()
		return scheduleOf(data, err, func() *entity.SchedJobSchedule { return &data.Schedule })
	}},
	{base.SettingTypeSSLRenewal, 1, 0, func(s *entity.Setting) (*entity.SchedJobSchedule, entity.SettingData, error) {
		data, err := s.AsSSLRenewal()
		return scheduleOf(data, err, func() *entity.SchedJobSchedule { return &data.Schedule })
	}},
	{base.SettingTypeBackupRepoCleanup, 1, 30, func(s *entity.Setting) (*entity.SchedJobSchedule, entity.SettingData,
		error) {
		data, err := s.AsBackupRepoCleanup()
		return scheduleOf(data, err, func() *entity.SchedJobSchedule { return &data.Schedule })
	}},
}

func scheduleOf(data entity.SettingData, err error, schedule func() *entity.SchedJobSchedule) (
	*entity.SchedJobSchedule, entity.SettingData, error) {
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return schedule(), data, nil
}

// dailyJobOf is the daily job of a setting type; the zero job for one that is
// none.
func dailyJobOf(settingType base.SettingType) dailyJob {
	for _, job := range dailyJobs {
		if job.settingType == settingType {
			return job
		}
	}
	return dailyJob{}
}

// startIn is when the job first runs, made now: at its time of day in loc, on
// the day it is there.
func (j dailyJob) startIn(now time.Time, loc *time.Location) time.Time {
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), j.hour, j.minute, 0, 0, loc).UTC()
}

// isDefaultIn says the schedule is still the one HivePaaS gave the job in loc:
// every 24 hours, from its time of day there.
func (j dailyJob) isDefaultIn(s *entity.SchedJobSchedule, loc *time.Location) bool {
	if s == nil || s.CronExpr != "" || s.Interval.ToDuration() != timeutil.Day || !s.EndTime.IsZero() {
		return false
	}
	start := s.InitialTime.In(loc)
	return start.Hour() == j.hour && start.Minute() == j.minute && start.Second() == 0 && start.Nanosecond() == 0
}

// movedTo is the schedule from the same day at the job's time of day in to.
// The runs made under the old one are forgotten, as a save of the schedule
// forgets them: the next are reckoned from the new start.
func (j dailyJob) movedTo(s *entity.SchedJobSchedule, from, to *time.Location) entity.SchedJobSchedule {
	day := s.InitialTime.In(from)
	return entity.SchedJobSchedule{
		Interval:    s.Interval,
		InitialTime: time.Date(day.Year(), day.Month(), day.Day(), j.hour, j.minute, 0, 0, to).UTC(),
	}
}

func (s *service) MoveDailyJobs(ctx context.Context, db database.Tx, from, to *time.Location) (
	moved []*entity.Setting, left []string, err error) {
	for _, job := range dailyJobs {
		jobSetting, name, err := s.moveDailyJob(ctx, db, job, from, to)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		switch {
		case jobSetting != nil:
			moved = append(moved, jobSetting)
		case name != "":
			left = append(left, name)
		}
	}
	return moved, left, nil
}

// moveDailyJob moves one daily job, its setting's schedule and its job's. It
// answers the job moved; or, for one left as it is, its setting's name.
func (s *service) moveDailyJob(ctx context.Context, db database.Tx, job dailyJob, from, to *time.Location) (
	*entity.Setting, string, error) {
	global := entity.NewObjectScopeGlobal()
	setting, err := s.settingRepo.GetSingle(ctx, db, global, job.settingType, false,
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, "", nil
		}
		return nil, "", hperrors.Wrap(err)
	}
	schedule, data, err := job.schedule(setting)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	if !job.isDefaultIn(schedule, from) {
		return nil, setting.Name, nil
	}
	*schedule = job.movedTo(schedule, from, to)
	if err = s.saveSetting(ctx, db, setting, data); err != nil {
		return nil, "", hperrors.Wrap(err)
	}

	jobSetting, err := s.settingRepo.GetSingle(ctx, db, global, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", setting.ID),
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, "", nil
		}
		return nil, "", hperrors.Wrap(err)
	}
	schedJob, err := jobSetting.AsSchedJob()
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	jobSchedule := *schedule
	schedJob.Schedule = &jobSchedule
	if err = s.saveSetting(ctx, db, jobSetting, schedJob); err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	return jobSetting, "", nil
}

func (s *service) saveSetting(ctx context.Context, db database.Tx, setting *entity.Setting,
	data entity.SettingData) error {
	if err := setting.SetData(data); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Update(ctx, db, setting))
}
