package projectservice

import (
	"context"
	"errors"

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
	rootProjectName = "localpaas"
	rootProjectKey  = "localpaas"
)

func (s *projectService) InitRootProject(
	ctx context.Context,
	db database.IDB,
) error {
	project, err := s.projectRepo.GetByKey(ctx, db, rootProjectKey)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if project == nil {
		timeNow := timeutil.NowUTC()
		project = &entity.Project{
			ID:        gofn.Must(ulid.NewStringULID()),
			Name:      rootProjectName,
			Key:       rootProjectKey,
			Status:    base.ProjectStatusActive,
			OwnerID:   "",
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}

		// Get admin account and assign it to project as owner
		users, _, err := s.userRepo.List(ctx, db, nil,
			bunex.SelectColumns("id"),
			bunex.SelectWhere("role = ?", base.UserRoleAdmin),
			bunex.SelectWhere("status = ?", base.UserStatusActive),
			bunex.SelectOrder("created_at"),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(users) > 0 {
			project.OwnerID = users[0].ID
		}
	}

	err = s.projectRepo.Upsert(ctx, db, project,
		entity.ProjectUpsertingConflictCols, entity.ProjectUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.SyncProject(ctx, db, project)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
