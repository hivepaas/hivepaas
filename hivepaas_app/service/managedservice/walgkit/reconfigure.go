package walgkit

import (
	"slices"
	"strconv"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// Reconfigure applies what every engine's settings share: the schedules, the
// retention, the restore test's check and where it runs, the dump, and the
// member the base backups are taken on. The setting is changed in place, and the
// jobs it rewrote, made or removes are returned.
func Reconfigure(p *Profile, req *managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	resp := &managedmodel.ReconfigureResp{}
	db, change := req.Managed.Database, req.Change
	if db == nil || change == nil {
		return resp, nil
	}
	for _, schedule := range []*string{change.BaseBackupSchedule, change.RestoreTestSchedule, change.DumpSchedule} {
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
		db.RetainFull = *change.RetainFull
		nightlyEnv[EnvRetainFull] = strconv.Itoa(db.RetainFull)
	}
	testEnv := map[string]string{}
	if change.RestoreTestQuery != nil {
		db.RestoreTest.CheckQuery = *change.RestoreTestQuery
		testEnv[EnvRestoreCheckQuery] = db.RestoreTest.CheckQuery
	}

	for _, update := range []struct {
		job      entity.ObjectID
		schedule *string
		env      map[string]string
	}{
		{db.Jobs.Nightly, change.BaseBackupSchedule, nightlyEnv},
		{db.Jobs.RestoreTest, change.RestoreTestSchedule, testEnv},
	} {
		setting := req.Jobs[update.job.ID]
		if setting == nil {
			continue
		}
		changed, rescheduled, err := RewriteJob(setting, update.schedule, update.env)
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
		setting := req.Jobs[db.Jobs.RestoreTest.ID]
		moved, err := moveRestoreTest(req.App, db, setting, change.RestoreTest)
		if err != nil {
			return nil, err
		}
		if moved && !slices.Contains(resp.Changed, setting) {
			resp.Changed = append(resp.Changed, setting)
		}
	}
	if err := reconfigureDump(p, req, resp); err != nil {
		return nil, err
	}
	if err := moveBaseBackups(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// moveBaseBackups points the nightly job at the member the base backups are
// taken on, or back at the database. Its retention runs on the database either
// way: the run takes its backup where the job points, and retains on the job's
// own app.
func moveBaseBackups(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	db, change := req.Managed.Database, req.Change
	if change.BackupFrom == nil {
		return nil
	}
	target := req.App
	if *change.BackupFrom != "" {
		if db.Member(*change.BackupFrom) == nil {
			return hperrors.NewArgumentInvalid("backupFrom").
				WithExtraDetail("%s is not a replica of this database", *change.BackupFrom)
		}
		db.BackupFrom = entity.ObjectID{ID: *change.BackupFrom}
	} else {
		db.BackupFrom = entity.ObjectID{}
	}
	if target == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("moving the base backups needs the app")
	}
	where := gofn.Coalesce(db.BackupFrom.ID, db.Writer.ID, target.ID)
	setting := req.Jobs[db.Jobs.Nightly.ID]
	if setting == nil {
		return nil
	}
	job, err := setting.AsSchedJob()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if job.App.ID == where {
		return nil
	}
	job.App = entity.ObjectID{ID: where}
	if err = setting.SetData(job); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	if !slices.Contains(resp.Changed, setting) {
		resp.Changed = append(resp.Changed, setting)
	}
	return nil
}

// moveRestoreTest points the restore test at the verifier given, or back at the
// database, and says whether its job changed. Its schedule stays.
func moveRestoreTest(app *entity.App, db *entity.ManagedDatabase, setting *entity.Setting,
	change *managedmodel.RestoreTestChange) (bool, error) {
	if app == nil {
		return false, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("moving a restore test needs its app")
	}
	moved := entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer, CheckQuery: db.RestoreTest.CheckQuery}
	jobType, target := base.SchedJobTypeContainerCommand, app.ID
	if change.Verifier != nil {
		moved.Mode = entity.ManagedRestoreTestVerifier
		moved.VerifierApp = entity.ObjectID{ID: change.Verifier.ID}
		moved.NodeID, moved.KeyAuth = change.NodeID, change.KeyAuth
		jobType, target = base.SchedJobTypeManagedRestoreTest, change.Verifier.ID
	}
	db.RestoreTest = moved
	if setting == nil {
		return false, nil
	}

	job, err := setting.AsSchedJob()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	changed := job.JobType != jobType || job.App.ID != target
	job.JobType, job.App = jobType, entity.ObjectID{ID: target}
	setting.Kind = string(jobType)
	if change.Verifier != nil {
		changed = SetCommandEnv(job.Command, EnvRestoreSourceHost, db.WriterKey(app)) || changed
		changed = SetCommandEnv(job.Command, EnvRestoreScratchDir, VerifierScratchDir) || changed
	} else if job.Command != nil {
		kept := slices.DeleteFunc(slices.Clone(job.Command.EnvVars), func(env *entity.EnvVar) bool {
			return env.Key == EnvRestoreSourceHost || env.Key == EnvRestoreScratchDir
		})
		changed = changed || len(kept) != len(job.Command.EnvVars)
		job.Command.EnvVars = kept
	}
	if !changed {
		return false, nil
	}
	if err = setting.SetData(job); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	return true, nil
}

// RewriteJob gives a job a schedule and variables of its command, and says
// whether that changed the job, and its schedule.
func RewriteJob(setting *entity.Setting, schedule *string, env map[string]string) (changed, rescheduled bool,
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
		if SetCommandEnv(job.Command, key, value) {
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

// SetCommandEnv sets a variable of a job's command, and says whether it changed.
func SetCommandEnv(command *entity.CommandTemplate, key, value string) bool {
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
