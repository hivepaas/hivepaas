package sysupdateserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

// afterUpdateTimeout bounds the restore-and-record step. It scales two services
// and sends the result notifications, so it has to allow for the network; what it
// must not do is run unbounded, since the update has already ended by then.
const afterUpdateTimeout = 5 * time.Minute

type sysUpdateData struct {
	*sysupdateservice.SysUpdateReq

	TaskOutput            *entity.TaskSystemUpdateOutput
	CurrentAppReplicas    *uint64
	CurrentWorkerReplicas *uint64

	NotifMsgData *notificationservice.TemplateDataSystemUpdate
}

func (s *service) SysUpdate(
	ctx context.Context,
	db database.IDB,
	req *sysupdateservice.SysUpdateReq,
) (resp *sysupdateservice.SysUpdateResp, err error) {
	resp = &sysupdateservice.SysUpdateResp{}
	data := &sysUpdateData{
		SysUpdateReq: req,
		TaskOutput:   &entity.TaskSystemUpdateOutput{},
	}

	defer func() {
		// A context of its own, and this is the whole point of it. Everything in
		// here puts the system back and records what happened, and it is needed
		// most when the update ended because ctx itself ran out: the task carries
		// a one hour deadline, and scaling the app back up through an expired
		// context does nothing at all, leaving an installation stopped with no
		// dashboard to fix it from.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), afterUpdateTimeout)
		defer cancel()

		// Finalize the update
		err2 := s.onAfterSystemUpdate(ctx, data)
		err = errors.Join(err, err2)

		// Update task fields
		task := data.Task
		task.EndedAt = timeutil.NowUTC()
		if err != nil {
			task.Status = base.TaskStatusFailed
			_ = task.AddRun(&entity.TaskRun{
				StartedAt: task.StartedAt,
				EndedAt:   task.EndedAt,
				Error:     hperrors.GetErrorDetail(err, ""),
			})
		} else {
			task.Status = base.TaskStatusDone
		}

		// Send result notifications
		s.sendResultNotifications(ctx, db, data)
	}()
	defer safego.RecoverTo(&err) // Early catch panic before the above defers

	// Before anything stops. Pulling is the one part of an update that does not
	// need the system to be down, and doing it here means a registry that cannot
	// be reached is found out while the app is still serving.
	if pullErr := s.pullAllImages(ctx, data); pullErr != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Some images could not be pulled ahead of time; swarm will fetch them as it needs them: "+
				pullErr.Error(), tasklog.TsNow))
	}

	// Stop only services which need to be stopped (main app and workers)
	err = s.stopServices(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.onBeforeSystemUpdate(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.updateSystem(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

func (s *service) stopServices(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	// 1. Scale down the main app to zero instance
	err = s.scaleMainAppService(ctx, 0, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 2. Scale down the workers to zero instance
	err = s.scaleWorkerService(ctx, 0, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) onBeforeSystemUpdate(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	// Dump the database, so a migration that does not complete can be undone.
	//
	// A failure here stops the update rather than warning and carrying on. The
	// dump is the only way back from a half-applied schema, and starting an
	// update without one is the risk this exists to remove - which is also why
	// skipping it has to be asked for explicitly.
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
	if args.SkipBackup {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Skipping the database backup, as requested - a failed migration cannot be undone",
			tasklog.TsNow))
		return nil
	}

	err = s.backupDB(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) onAfterSystemUpdate(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	// Bring back the main app instances
	if data.CurrentAppReplicas != nil && *data.CurrentAppReplicas > 0 {
		err = s.scaleMainAppService(ctx, *data.CurrentAppReplicas, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Bring back the worker instances
	if data.CurrentWorkerReplicas != nil && *data.CurrentWorkerReplicas > 0 {
		err = s.scaleWorkerService(ctx, *data.CurrentWorkerReplicas, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) updateSystem(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	defer safego.RecoverTo(&err)

	// 1. Update DB
	err = s.updateDbService(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 2. Update redis
	err = s.updateRedisService(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 3. Update traefik
	err = s.updateTraefikService(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 4. Update the logging stack, if it is deployed at all. Before the app and
	// the worker only because those two are what bring the system back up, and
	// nothing here is a dependency of either.
	err = s.updateLoggingService(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 5. The app and the worker last, and together.
	//
	// They are what brings the system back, they run the same image, and each
	// waits out its own 240s update monitor - so doing them one after the other
	// spends four minutes twice for no reason. Nothing else in the update depends
	// on either, and onAfterSystemUpdate already treats them as a pair.
	//
	// Their log lines interleave, which is readable only because every line one
	// of these steps writes is prefixed with the service it is about.
	//
	// stopOnError is off: abandoning the worker halfway because the app failed
	// saves nothing, and both failures are worth reporting.
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
	errMap := gofn.ExecTasksEx(ctx, 0, false,
		func(ctx context.Context) error {
			return s.updateMainAppService(ctx, data, args)
		},
		func(ctx context.Context) error {
			return s.updateWorkerService(ctx, data, args)
		},
	)

	updateErrs := make([]error, 0, len(errMap))
	for _, updateErr := range errMap {
		updateErrs = append(updateErrs, updateErr)
	}
	return hperrors.Wrap(errors.Join(updateErrs...))
}

func (s *service) sendResultNotifications(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) {
	task := data.Task
	if task.IsDone() || task.IsFailedCompletely() {
		err := s.notifyForSystemUpdate(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx,
				tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
				tasklog.NewOutFrame("Failed to send system update notification with error: "+err.Error(),
					tasklog.TsNow))
		}
	}
}
