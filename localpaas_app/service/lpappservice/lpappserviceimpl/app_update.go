package lpappserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
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
		return apperrors.New(apperrors.ErrTooMany).WithParam("Name", "Update requests")
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
		CurrentVersion: base.StableVersion, // TODO: correct this
		TargetVersion:  targetVersion,
	})

	err = s.taskRepo.Insert(ctx, db, task)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Start the updater service
	updaterSvc, err := s.GetLpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	appSvc, err := s.GetLpAppSwarmService(ctx)
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
