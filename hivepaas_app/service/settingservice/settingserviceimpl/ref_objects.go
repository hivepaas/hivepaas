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
	scope *entity.ObjectScope,
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
	scope *entity.ObjectScope,
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
			return nil, apperrors.Wrap(err)
		}
	}

	// Load ref apps
	if len(refIDs.RefAppIDs) > 0 {
		refObjects.RefApps, err = s.LoadReferenceApps(ctx, db, requireActive, errorIfUnavail,
			refIDs.RefAppIDs)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// Load ref projects
	if len(refIDs.RefProjectIDs) > 0 {
		refObjects.RefProjects, err = s.LoadReferenceProjects(ctx, db, requireActive, errorIfUnavail,
			refIDs.RefProjectIDs)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// Load ref project envs
	if len(refIDs.RefProjectEnvIDs) > 0 {
		refObjects.RefProjectEnvs, err = s.LoadReferenceProjectEnvs(ctx, db, requireActive, errorIfUnavail,
			refIDs.RefProjectEnvIDs)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// Load ref settings
	if len(refIDs.RefSettingIDs) > 0 {
		refObjects.RefSettings, err = s.LoadReferenceSettings(ctx, db, scope, requireActive,
			errorIfUnavail, refIDs.RefSettingIDs)
		if err != nil {
			return nil, apperrors.Wrap(err)
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
		return nil, apperrors.Wrap(err)
	}
	refObjects.AddRefObjects(newRecursiveRefObjects)

	return refObjects, nil
}

func (s *service) LoadReferenceSettings(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
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
		return nil, apperrors.Wrap(err)
	}
	settingMap = entityutil.SliceToIDMap(settings)

	// Check setting availability
	if errorIfUnavail {
		for _, id := range settingIDs {
			if _, exists := settingMap[id]; !exists {
				return nil, apperrors.Wrap(apperrors.ErrSettingNotFound).WithParam("Name", id).
					WithMsgLog("setting %s not found or expired", id)
			}
		}
	}

	return settingMap, nil
}

func (s *service) LoadReferenceApps(
	ctx context.Context,
	db database.IDB,
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
		bunex.SelectRelation("ProjectEnv"),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("app.status = ?", base.AppStatusActive))
	}

	apps, err := s.appRepo.ListByIDs(ctx, db, "", appIDs, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	appMap = entityutil.SliceToIDMap(apps)

	for _, id := range appIDs {
		app := appMap[id]
		if (requireActive || errorIfUnavail) && app == nil {
			return nil, apperrors.Wrap(apperrors.ErrAppNotFound).WithParam("Name", id)
		}
		if app == nil {
			continue
		}
		if errorIfUnavail && app.Project == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).
				WithParam("Name", app.ProjectID)
		}
		if errorIfUnavail && app.ProjectEnv == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectEnvNotFound).
				WithParam("Name", app.ProjectEnvID)
		}
		if requireActive && (app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
			return nil, apperrors.Wrap(apperrors.ErrProjectInactive).
				WithParam("Name", app.ProjectID)
		}
		if requireActive && app.ProjectEnv != nil && app.ProjectEnv.Status != base.ProjectStatusActive {
			return nil, apperrors.Wrap(apperrors.ErrProjectEnvInactive).
				WithParam("Name", app.ProjectEnvID)
		}
	}

	return appMap, nil
}

func (s *service) LoadReferenceProjects(
	ctx context.Context,
	db database.IDB,
	requireActive bool,
	errorIfUnavail bool,
	projectIDs []string,
) (projectMap map[string]*entity.Project, err error) {
	projectIDs = gofn.ToSet(projectIDs)
	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("project.status = ?", base.ProjectStatusActive))
	}

	projects, err := s.projectRepo.ListByIDs(ctx, db, projectIDs, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	projectMap = entityutil.SliceToIDMap(projects)

	for _, id := range projectIDs {
		project := projectMap[id]
		if (requireActive || errorIfUnavail) && project == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", id)
		}
	}

	return projectMap, nil
}

func (s *service) LoadReferenceProjectEnvs(
	ctx context.Context,
	db database.IDB,
	requireActive bool,
	errorIfUnavail bool,
	projectEnvIDs []string,
) (projectEnvMap map[string]*entity.ProjectEnv, err error) {
	projectEnvIDs = gofn.ToSet(projectEnvIDs)
	opts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("project_env.status = ?", base.ProjectStatusActive))
	}

	projectEnvs, err := s.projectEnvRepo.ListByIDs(ctx, db, projectEnvIDs, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	projectEnvMap = entityutil.SliceToIDMap(projectEnvs)

	for _, id := range projectEnvIDs {
		projectEnv := projectEnvMap[id]
		if (requireActive || errorIfUnavail) && projectEnv == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", id)
		}
		if projectEnv == nil {
			continue
		}
		if errorIfUnavail && projectEnv.Project == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).
				WithParam("Name", projectEnv.ProjectID)
		}
		if requireActive && (projectEnv.Project == nil || projectEnv.Project.Status != base.ProjectStatusActive) {
			return nil, apperrors.Wrap(apperrors.ErrProjectInactive).
				WithParam("Name", projectEnv.ProjectID)
		}
	}

	return projectEnvMap, nil
}
