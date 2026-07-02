package projectserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) InitRootProject(
	ctx context.Context,
	db database.IDB,
) (postInitFunc func() error, err error) {
	project, err := s.projectRepo.GetByKey(ctx, db, base.HivepaasProjectKey)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New(err)
	}
	if project == nil {
		timeNow := timeutil.NowUTC()
		project = &entity.Project{
			ID:        gofn.Must(ulid.NewStringULID()),
			Name:      base.HivepaasProjectName,
			Key:       base.HivepaasProjectKey,
			Status:    base.ProjectStatusActive,
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
			return nil, apperrors.New(err)
		}
		if len(users) > 0 {
			project.OwnerID = users[0].ID
		}
	}

	err = s.projectRepo.Upsert(ctx, db, project,
		entity.ProjectUpsertingConflictCols, entity.ProjectUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.New(err)
	}

	newApps, _, services, err := s.SyncProject(ctx, db, project)
	if err != nil {
		return nil, apperrors.New(err)
	}

	var updatingServices []*swarm.Service
	for _, app := range newApps {
		var svc *swarm.Service
		for i := range services {
			if services[i].ID == app.ServiceID {
				svc = &services[i]
				break
			}
		}
		shouldUpdateService := false
		switch app.Key {
		case base.HivepaasAppKey:
			shouldUpdateService, err = s.initRootProjectMainApp(ctx, db, app, svc)
		case base.HivepaasTraefikAppKey:
			shouldUpdateService, err = s.initRootProjectTraefikApp(ctx, db, app, svc)
		}
		if err != nil {
			return nil, apperrors.New(err)
		}
		if shouldUpdateService {
			updatingServices = append(updatingServices, svc)
		}
	}

	postInitFunc = func() error {
		for _, svc := range updatingServices {
			err := gofn.ExecRetry(func() error {
				_, err := s.dockerManager.ServiceUpdate(ctx, svc.ID, &svc.Version, &svc.Spec)
				return apperrors.New(err)
			}, 2, time.Second*5) //nolint
			if err != nil {
				return apperrors.New(err)
			}
		}
		return nil
	}

	return postInitFunc, nil
}
