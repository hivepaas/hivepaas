package schedjobtriggerserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

func (s *service) Fire(
	ctx context.Context,
	event base.SchedJobTriggerEvent,
	app *entity.App,
	info *schedjobtriggerservice.TriggerInfo,
) (*schedjobtriggerservice.FireResult, error) {
	result := &schedjobtriggerservice.FireResult{}
	if app == nil {
		return result, nil
	}
	// An app and its env hold few jobs: they are listed whole and matched here.
	jobs, _, err := s.settingRepo.List(ctx, s.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSchedJob),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id IN (?)", bunex.List([]string{app.ID, app.ProjectEnvID})),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	featureOn, err := s.schedJobFeatureOn(ctx, s.db, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	result.Runs, err = s.buildRuns(app, jobs, event, info, featureOn, timeutil.NowUTC())
	if err != nil || len(result.Runs) == 0 {
		return result, hperrors.Wrap(err)
	}

	tasks := make([]*entity.Task, 0, len(result.Runs))
	for _, run := range result.Runs {
		tasks = append(tasks, run.Task)
	}
	err = transaction.Execute(ctx, s.db, func(tx database.Tx) error {
		return hperrors.Wrap(s.taskRepo.InsertMulti(ctx, tx, tasks))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.taskQueue.ScheduleTask(ctx, tasks...); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return result, nil
}

// buildRuns is a run of every job listening to event of app: the app's own
// jobs, when its scheduled jobs feature is on, and the env's jobs naming it.
func (s *service) buildRuns(
	app *entity.App,
	jobs []*entity.Setting,
	event base.SchedJobTriggerEvent,
	info *schedjobtriggerservice.TriggerInfo,
	featureOn bool,
	timeNow time.Time,
) ([]*schedjobtriggerservice.Run, error) {
	var runs []*schedjobtriggerservice.Run
	for _, setting := range jobs {
		if !setting.IsActive() || setting.Type != base.SettingTypeSchedJob {
			continue
		}
		ownApp := setting.Scope == base.ObjectScopeApp && setting.ObjectID == app.ID
		if ownApp && !featureOn {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		listens, wait := job.ListensTo(event, app.ID, ownApp)
		if !listens {
			continue
		}
		task, err := s.schedJobService.CreateSchedJobTask(setting, timeNow, timeNow)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		cause := &entity.SchedJobTriggerCause{Event: event, AppID: app.ID}
		if info != nil {
			cause.DeploymentID = info.DeploymentID
		}
		task.MustSetArgs(&entity.TaskSchedJobExecArgs{Trigger: cause})
		runs = append(runs, &schedjobtriggerservice.Run{
			Task:    task,
			JobName: setting.Name,
			Wait:    wait,
			Timeout: waitTimeout(job),
		})
	}
	return runs, nil
}

// waitTimeout is how long a deploy waits for a run of job: its own timeout, or
// the configured one for a job without a timeout and for a job sequence, whose
// timeout is each step's.
func waitTimeout(job *entity.SchedJob) time.Duration {
	if job.Timeout > 0 && job.JobType != base.SchedJobTypeJobSequence {
		return job.Timeout.ToDuration()
	}
	if cfg := config.Current(); cfg != nil && cfg.Tasks.Triggers.WaitTimeout > 0 {
		return cfg.Tasks.Triggers.WaitTimeout
	}
	return defaultWaitTimeout
}

const defaultWaitTimeout = 30 * time.Minute

// schedJobFeatureOn says whether the app's scheduled jobs feature is on; it is
// by default.
func (s *service) schedJobFeatureOn(ctx context.Context, db database.IDB, app *entity.App) (bool, error) {
	featureSetting, err := s.settingRepo.GetSingle(ctx, db, app.GetObjectScope(), base.SettingTypeAppFeatures, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return false, hperrors.Wrap(err)
	}
	features := &entity.AppFeatureSettings{}
	if featureSetting != nil {
		features = featureSetting.MustAsAppFeatureSettings()
	} else {
		entity.InitAppFeatureSettingsDefault(features)
	}
	return features.SchedJobSettings == nil || features.SchedJobSettings.Enabled, nil
}
