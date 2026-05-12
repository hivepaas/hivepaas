package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sysupdateservice"
)

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
				Error:     err.Error(),
			})
		} else {
			task.Status = base.TaskStatusDone
		}

		// Send result notifications (not count its error if happens)
		_ = s.sendResultNotifications(ctx, db, data)
	}()
	defer funcutil.EnsureNoPanic(&err) // Early catch panic before the above defers

	// Stop only services which need to be stopped (main app and workers)
	err = s.stopServices(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.onBeforeSystemUpdate(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.updateSystem(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}

	// 2. Scale down the workers to zero instance
	err = s.scaleWorkerService(ctx, 0, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) onBeforeSystemUpdate(
	_ context.Context,
	_ *sysUpdateData,
) (err error) {
	// 1. Pull all images we need
	// err = e.pullAllImages(ctx, data)
	// if err != nil {
	//	return apperrors.Wrap(err)
	// }

	// TODO: backup DB data

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
			return apperrors.Wrap(err)
		}
	}

	// Bring back the worker instances
	if data.CurrentWorkerReplicas != nil && *data.CurrentWorkerReplicas > 0 {
		err = s.scaleWorkerService(ctx, *data.CurrentWorkerReplicas, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) updateSystem(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	defer funcutil.EnsureNoPanic(&err)

	// 1. Update DB
	err = s.updateDbService(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 2. Update redis
	err = s.updateRedisService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 3. Update traefik
	err = s.updateTraefikService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Update main app then bring it back
	err = s.updateMainAppService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Update worker then bring it back
	err = s.updateWorkerService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return err
}

func (s *service) sendResultNotifications(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) error {
	task := data.Task
	if task.IsDone() || task.IsFailedCompletely() {
		err := s.notifyForSystemUpdate(ctx, db, data)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to send system update notification"+
				" with error: "+err.Error(), applog.TsNow))
			return apperrors.Wrap(err)
		}
	}
	return nil
}
