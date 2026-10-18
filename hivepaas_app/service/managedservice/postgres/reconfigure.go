package postgres

import (
	"slices"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func (e *Engine) Reconfigure(req *managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	resp := &managedmodel.ReconfigureResp{}
	pg, change := req.Managed.Postgres, req.Change
	if pg == nil || change == nil {
		return resp, nil
	}
	for _, schedule := range []*string{change.BaseBackupSchedule, change.RestoreTestSchedule} {
		if schedule == nil {
			continue
		}
		if _, err := (&entity.SchedJobSchedule{CronExpr: *schedule}).ParseCronExpr(); err != nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
				WithExtraDetail("%q is not a cron expression", *schedule)
		}
	}

	nightlyEnv := map[string]string{}
	if change.RetainFull != nil {
		pg.RetainFull = *change.RetainFull
		nightlyEnv[envRetainFull] = strconv.Itoa(pg.RetainFull)
	}
	testEnv := map[string]string{}
	if change.RestoreTestQuery != nil {
		pg.RestoreTest.CheckQuery = *change.RestoreTestQuery
		testEnv[envRestoreCheckQuery] = pg.RestoreTest.CheckQuery
	}

	for _, update := range []struct {
		job      entity.ObjectID
		schedule *string
		env      map[string]string
	}{
		{pg.Jobs.Nightly, change.BaseBackupSchedule, nightlyEnv},
		{pg.Jobs.RestoreTest, change.RestoreTestSchedule, testEnv},
	} {
		setting := req.Jobs[update.job.ID]
		if setting == nil {
			continue
		}
		changed, rescheduled, err := rewriteJob(setting, update.schedule, update.env)
		if err != nil {
			return nil, err
		}
		if changed {
			resp.Changed = append(resp.Changed, setting)
		}
		if rescheduled {
			resp.Rescheduled = append(resp.Rescheduled, setting)
		}
	}

	if change.RestoreTest != nil {
		setting := req.Jobs[pg.Jobs.RestoreTest.ID]
		moved, err := moveRestoreTest(req.App, pg, setting, change.RestoreTest)
		if err != nil {
			return nil, err
		}
		if moved && !slices.Contains(resp.Changed, setting) {
			resp.Changed = append(resp.Changed, setting)
		}
	}
	return resp, nil
}

// rewriteJob gives a job a schedule and variables of its command, and says
// whether that changed the job, and its schedule.
func rewriteJob(setting *entity.Setting, schedule *string, env map[string]string) (changed, rescheduled bool,
	err error) {
	job, err := setting.AsSchedJob()
	if err != nil {
		return false, false, hperrors.Wrap(err)
	}
	if schedule != nil && job.Schedule != nil && job.Schedule.CronExpr != *schedule {
		job.Schedule.CronExpr = *schedule
		changed, rescheduled = true, true
	}
	for key, value := range env {
		if setCommandEnv(job.Command, key, value) {
			changed = true
		}
	}
	if !changed {
		return false, false, nil
	}
	if err = setting.SetData(job); err != nil {
		return false, false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	return changed, rescheduled, nil
}

// setCommandEnv sets a variable of a job's command, and says whether it changed.
func setCommandEnv(command *entity.CommandTemplate, key, value string) bool {
	if command == nil {
		return false
	}
	for _, env := range command.EnvVars {
		if env.Key == key {
			if env.Value == value {
				return false
			}
			env.Value = value
			return true
		}
	}
	command.EnvVars = append(command.EnvVars, &entity.EnvVar{Key: key, Value: value})
	return true
}
