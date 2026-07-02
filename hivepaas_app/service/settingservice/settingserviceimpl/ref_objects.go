package settingserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
)

func (s *service) LoadReferenceObjects(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	inSettings ...*entity.Setting,
) (refObjects *entity.RefObjects, err error) {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.LoadReferenceObjectsByIDs(ctx, db, scope, requireActive, errorIfUnavail, allRefIDs)
}

func (s *service) LoadReferenceObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	refIDs *entity.RefObjectIDs,
) (refObjects *entity.RefObjects, err error) {
	refObjects = entity.NewRefObjects()

	if refIDs == nil || !refIDs.HasData() {
		return refObjects, nil
	}

	// Load ref users
	if len(refIDs.RefUserIDs) > 0 {
		refObjects.RefUsers, err = s.userService.LoadUsers(ctx, db, refIDs.RefUserIDs, errorIfUnavail)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	// Make sure the current app id is in the list
	if scope != nil && scope.IsAppScope() && !gofn.Contain(refIDs.RefAppIDs, scope.AppID) {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, scope.AppID)
	}
	// Load ref apps
	if len(refIDs.RefAppIDs) > 0 {
		projectID := ""
		if scope != nil {
			projectID = scope.ProjectID
		}
		refObjects.RefApps, err = s.LoadReferenceApps(ctx, db, projectID, requireActive,
			errorIfUnavail, refIDs.RefAppIDs)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	// Load ref settings
	if len(refIDs.RefSettingIDs) > 0 {
		refObjects.RefSettings, err = s.LoadReferenceSettings(ctx, db, scope, requireActive,
			errorIfUnavail, refIDs.RefSettingIDs)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	// Calculate recursive ref IDs to load
	newRecursiveRefIDs := refIDs.GetRecursiveRefObjectIDs(refObjects)
	if !newRecursiveRefIDs.HasData() {
		return refObjects, nil
	}

	newRecursiveRefObjects, err := s.LoadReferenceObjectsByIDs(ctx, db, scope, requireActive,
		errorIfUnavail, newRecursiveRefIDs)
	if err != nil {
		return nil, apperrors.New(err)
	}
	refObjects.AddRefObjects(newRecursiveRefObjects)

	return refObjects, nil
}

func (s *service) LoadReferenceSettings(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	settingIDs []string,
) (settingMap map[string]*entity.Setting, err error) {
	settingIDs = gofn.ToSet(settingIDs)
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhereIn("setting.id IN (?)", settingIDs...),
	}
	if requireActive {
		listOpts = append(listOpts, bunex.SelectWhere("setting.status = ?", base.SettingStatusActive))
	}

	settings, _, err := s.settingRepo.List(ctx, db, scope, nil, listOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	settingMap = entityutil.SliceToIDMap(settings)

	// Check setting availability
	if errorIfUnavail {
		for _, id := range settingIDs {
			if _, exists := settingMap[id]; !exists {
				return nil, apperrors.NewNotFound("Setting").
					WithMsgLog("setting %s not found or expired", id)
			}
		}
	}

	return settingMap, nil
}

func (s *service) LoadReferenceApps(
	ctx context.Context,
	db database.IDB,
	projectID string,
	requireActive bool,
	errorIfUnavail bool,
	appIDs []string,
) (appMap map[string]*entity.App, err error) {
	appIDs = gofn.ToSet(appIDs)
	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("app.status = ?", base.AppStatusActive))
	}

	apps, err := s.appRepo.ListByIDs(ctx, db, projectID, appIDs, opts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	appMap = entityutil.SliceToIDMap(apps)

	for _, id := range appIDs {
		app, exists := appMap[id]
		if errorIfUnavail && !exists {
			return nil, apperrors.NewNotFound("App").
				WithMsgLog("app %s not found or inactive", id)
		}
		if app == nil {
			continue
		}
		if requireActive && app.Project != nil && app.Project.Status != base.ProjectStatusActive {
			app.Project = nil
		}
		if errorIfUnavail && app.Project == nil {
			return nil, apperrors.NewNotFound("Project").
				WithMsgLog("project %s not found", app.ProjectID)
		}
	}

	return appMap, nil
}
