package hpappserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	defaultSystemUpdateTimeout = time.Minute * 60
)

func (s *service) UpdateSystemVersion(
	ctx context.Context,
	db database.IDB,
	targetVersion *base.ReleaseInfo,
) error {
	timeNow := timeutil.NowUTC()

	// Make sure there is no other update task in the system
	tasks, _, err := s.taskRepo.List(ctx, db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSystemUpdate),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectWhere("task.created_at > ?", timeNow.Add(-time.Hour)),
		bunex.SelectLimit(1),
		bunex.SelectColumns("id"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if len(tasks) > 0 {
		return apperrors.Wrap(apperrors.ErrTooMany).WithParam("Name", "Update requests").
			WithNTParam("MaxItem", 1)
	}

	// Create a task for the system update
	task := &entity.Task{
		ID:     gofn.Must(ulid.NewStringULID()),
		Type:   base.TaskTypeSystemUpdate,
		Status: base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Timeout: timeutil.Duration(defaultSystemUpdateTimeout),
		},
		Version:   entity.CurrentTaskVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	task.MustSetArgs(&entity.TaskSystemUpdateArgs{
		CurrentVersion: gofn.If(config.Current.IsBetaEnv(), base.BetaVersion, base.StableVersion),
		TargetVersion:  targetVersion,
	})

	err = s.taskRepo.Insert(ctx, db, task)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Start the updater service
	updaterSvc, err := s.GetHpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	appSvc, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	updaterSvc.Spec.TaskTemplate.ContainerSpec.Image = targetVersion.AppImage
	// Make sure the admin service has the same storages as the main service
	updaterSvc.Spec.TaskTemplate.ContainerSpec.Mounts = appSvc.Spec.TaskTemplate.ContainerSpec.Mounts
	// Turn on the updater service
	updaterSvc.Spec.Mode.Replicated.Replicas = new(uint64(1))

	_, err = s.dockerManager.ServiceUpdate(ctx, updaterSvc.ID, &updaterSvc.Version, &updaterSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
