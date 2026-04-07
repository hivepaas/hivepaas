package projectservice

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

func (s *projectService) SyncProject(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) error {
	// Loads all apps in project
	apps, _, err := s.appRepo.List(ctx, db, project.ID, nil)
	if err != nil {
		return apperrors.Wrap(err)
	}

	appMapByKey := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appMapByKey[app.Key] = app
	}

	// Loads all swarm services from docker having the namespace label as project key
	services, err := s.dockerManager.ServiceListByStack(ctx, project.Key)
	if err != nil {
		return apperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	var newApps []*entity.App
	var updateApps []*entity.App

	// Sync the services with the apps, create new apps if need to
	for _, svc := range services {
		appKey := svc.Spec.Name
		appName := strings.Trim(strings.TrimPrefix(appKey, project.Key), "_-")

		if existingApp, exists := appMapByKey[appKey]; exists {
			if existingApp.ServiceID != svc.ID {
				existingApp.ServiceID = svc.ID
				existingApp.UpdateVer++
				existingApp.UpdatedAt = timeNow
				updateApps = append(updateApps, existingApp)
			}
		} else {
			newApp := &entity.App{
				ID:        gofn.Must(ulid.NewStringULID()),
				Name:      appName,
				Key:       appKey,
				ProjectID: project.ID,
				ServiceID: svc.ID,
				Status:    base.AppStatusActive,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			}
			newApp.ResetToken()
			newApps = append(newApps, newApp)
		}
	}

	err = s.appRepo.UpsertMulti(ctx, db, gofn.Concat(newApps, updateApps),
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
