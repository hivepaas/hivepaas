package projectserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

func (s *service) SyncProject(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) (newApps, updateApps []*entity.App, services []swarm.Service, err error) {
	// Loads all apps in project
	apps, _, err := s.appRepo.List(ctx, db, project.ID, nil)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	appMapByKey := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appMapByKey[app.Key] = app
	}

	// Loads all swarm services from docker having the namespace label as project key
	listResp, err := s.dockerManager.ServiceListByStack(ctx, project.Key)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}
	services = listResp.Items

	timeNow := timeutil.NowUTC()

	// Sync the services with the apps, create new apps if need to
	for _, svc := range services {
		appInfoLabel := svc.Spec.Labels[appservice.LabelAppInfo]
		appInfo, err := apphelper.ParseAppInfoLabel(appInfoLabel)
		if err != nil {
			return nil, nil, nil, apperrors.Wrap(err)
		}

		appKey := appInfo.Key
		appName := gofn.Coalesce(appInfo.Name, svc.Spec.Name)
		if appKey == "" {
			appKey = slugify.SlugifyAsKey(appName)
		}
		appEnv := gofn.Coalesce(appInfo.Env, "default")
		appGlobalKey := projecthelper.CalcAppGlobalKey(project.Key, appKey, appEnv)

		if existingApp, exists := appMapByKey[appKey]; exists {
			if existingApp.ServiceID != svc.ID {
				existingApp.ServiceID = svc.ID
				existingApp.UpdateVer++
				existingApp.UpdatedAt = timeNow
				updateApps = append(updateApps, existingApp)
			}
		} else {
			projectEnv := project.GetOrCreateEnv(appEnv)
			newApp := &entity.App{
				ID:           gofn.Must(ulid.NewStringULID()),
				ProjectID:    project.ID,
				ProjectEnvID: projectEnv.ID,
				Name:         appName,
				Key:          appKey,
				GlobalKey:    appGlobalKey,
				ServiceID:    svc.ID,
				Status:       base.AppStatusActive,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			}
			newApps = append(newApps, newApp)
		}
	}

	err = s.projectEnvRepo.UpsertMulti(ctx, db, project.ProjectEnvs,
		entity.ProjectEnvUpsertingConflictCols, entity.ProjectEnvUpsertingUpdateCols)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	err = s.appRepo.UpsertMulti(ctx, db, gofn.Concat(newApps, updateApps),
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err)
	}

	return newApps, updateApps, services, nil
}
