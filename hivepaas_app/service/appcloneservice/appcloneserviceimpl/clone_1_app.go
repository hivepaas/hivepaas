package appcloneserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) cloneApp(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	timeNow := timeutil.NowUTC()
	srcApp := data.SrcApp
	destApp := &entity.App{
		ID:           gofn.Must(ulid.NewStringULID()),
		ProjectID:    srcApp.Project.ID,
		Project:      srcApp.Project,
		ProjectEnvID: srcApp.ProjectEnvID,
		ProjectEnv:   srcApp.ProjectEnv,
		Status:       base.AppStatusActive,
		CreatedAt:    timeNow,
		UpdatedAt:    timeNow,
	}
	data.DestApp = destApp

	cloneFunc := data.OnCloneApp
	if cloneFunc == nil {
		cloneFunc = func(destApp, srcApp *entity.App) error {
			return s.onCloneAppDefault(destApp, srcApp, data)
		}
	}

	err = cloneFunc(destApp, srcApp)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if destApp.ProjectEnv == nil {
		destApp.ProjectEnv = srcApp.ProjectEnv
	}
	destEnv := destApp.ProjectEnv.Name
	destApp.Key = projecthelper.CalcAppKey(destApp.Name)
	if destApp.ParentApp != nil {
		destApp.Key = destApp.ParentApp.Key + "_" + destApp.Key
	}
	destApp.GlobalKey = projecthelper.CalcAppGlobalKey(destApp.Project.Key, destApp.Key, destEnv)

	// App global key must be unique globally
	conflictApp, err := s.appRepo.GetByGlobalKey(ctx, db, "", destApp.GlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if conflictApp != nil {
		return apperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", destApp.GlobalKey)
	}

	// Create local network for the app to attach
	_, _, err = s.networkService.GetOrCreateProjectNetwork(ctx, db, destApp.Project, destEnv)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) onCloneAppDefault(
	destApp, srcApp *entity.App,
	data *appCloneData,
) error {
	settings := data.CloneSettings
	// Name
	destApp.Name = settings.TargetName
	if destApp.Name == "" {
		destApp.Name = srcApp.Name + " (clone)"
	}

	// Status
	destApp.Status = gofn.Coalesce(settings.TargetStatus, base.AppStatusActive)

	// Env
	destEnv := settings.TargetEnv
	if destEnv == "" {
		destEnv = srcApp.ProjectEnv.Key
	}
	projectEnv := srcApp.Project.GetEnv(destEnv)
	if projectEnv == nil {
		return apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", destEnv)
	}
	destApp.ProjectEnvID = projectEnv.ID
	destApp.ProjectEnv = projectEnv

	return nil
}
