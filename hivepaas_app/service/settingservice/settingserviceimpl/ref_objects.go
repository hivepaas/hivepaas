package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadRefObjects(
	ctx context.Context,
	db database.IDB,
	refObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	inSettings ...*entity.Setting,
) (err error) {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.LoadRefObjectsByIDs(ctx, db, refObjects, scope, requireActive, errorIfUnavail, allRefIDs)
}

func (s *service) LoadRefObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	refIDs *entity.RefObjectIDs,
) (err error) {
	if pRefObjects == nil {
		return apperrors.NewArgumentInvalid("refObjects")
	}
	if *pRefObjects == nil {
		*pRefObjects = entity.NewRefObjects()
	}
	refObjects := *pRefObjects

	if refIDs == nil || !refIDs.HasData() {
		return nil
	}

	// Load ref users
	if len(refIDs.RefUserIDs) > 0 {
		err = s.loadReferenceUsers(ctx, db, pRefObjects, errorIfUnavail, refIDs.RefUserIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref apps
	if len(refIDs.RefAppIDs) > 0 {
		err = s.loadReferenceApps(ctx, db, pRefObjects, requireActive, errorIfUnavail, refIDs.RefAppIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref projects
	if len(refIDs.RefProjectIDs) > 0 {
		err = s.loadReferenceProjects(ctx, db, pRefObjects, requireActive, errorIfUnavail, refIDs.RefProjectIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref project envs
	if len(refIDs.RefProjectEnvIDs) > 0 {
		err = s.loadReferenceProjectEnvs(ctx, db, pRefObjects, requireActive, errorIfUnavail, refIDs.RefProjectEnvIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref settings
	if len(refIDs.RefSettingIDs) > 0 {
		err = s.loadReferenceSettings(ctx, db, pRefObjects, scope, requireActive, errorIfUnavail, refIDs.RefSettingIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Calculate recursive ref IDs to load
	newRecursiveRefIDs := refIDs.GetRecursiveRefObjectIDs(refObjects)
	if !newRecursiveRefIDs.HasData() {
		return nil
	}

	err = s.LoadRefObjectsByIDs(ctx, db, pRefObjects, scope, requireActive, errorIfUnavail, newRecursiveRefIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) loadReferenceSettings(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	errorIfUnavail bool,
	settingIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadSettingIDs := make([]string, 0, len(settingIDs))
	for _, settingID := range settingIDs {
		if _, ok := refObjects.RefSettings[settingID]; ok {
			continue
		}
		loadSettingIDs = append(loadSettingIDs, settingID)
	}
	if len(loadSettingIDs) == 0 {
		return nil
	}

	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhereIn("setting.id IN (?)", loadSettingIDs...),
	}
	if requireActive {
		listOpts = append(listOpts, bunex.SelectWhere("setting.status = ?", base.SettingStatusActive))
	}

	settings, _, err := s.settingRepo.List(ctx, db, scope, nil, listOpts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, setting := range settings {
		refObjects.RefSettings[setting.ID] = setting
	}

	// Check setting availability
	if errorIfUnavail {
		for _, id := range loadSettingIDs {
			if _, exists := refObjects.RefSettings[id]; !exists {
				return apperrors.Wrap(apperrors.ErrSettingNotFound).WithParam("Name", id).
					WithMsgLog("setting %s not found or expired", id)
			}
		}
	}

	return nil
}

func (s *service) loadReferenceApps(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireActive bool,
	errorIfUnavail bool,
	appIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadAppIDs := make([]string, 0, len(appIDs))
	for _, appID := range appIDs {
		if app, ok := refObjects.RefApps[appID]; ok && app.Project != nil && app.ProjectEnv != nil {
			continue
		}
		loadAppIDs = append(loadAppIDs, appID)
	}
	if len(loadAppIDs) == 0 {
		return nil
	}

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

	apps, err := s.appRepo.ListByIDs(ctx, db, "", loadAppIDs, opts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, app := range apps {
		refObjects.RefApps[app.ID] = app
	}

	for _, id := range loadAppIDs {
		app := refObjects.RefApps[id]
		if (requireActive || errorIfUnavail) && app == nil {
			return apperrors.Wrap(apperrors.ErrAppNotFound).WithParam("Name", id)
		}
		if app == nil {
			continue
		}
		if errorIfUnavail && app.Project == nil {
			return apperrors.Wrap(apperrors.ErrProjectNotFound).
				WithParam("Name", app.ProjectID)
		}
		if errorIfUnavail && app.ProjectEnv == nil {
			return apperrors.Wrap(apperrors.ErrProjectEnvNotFound).
				WithParam("Name", app.ProjectEnvID)
		}
		if requireActive && (app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
			return apperrors.Wrap(apperrors.ErrProjectInactive).
				WithParam("Name", app.ProjectID)
		}
		if requireActive && app.ProjectEnv != nil && app.ProjectEnv.Status != base.ProjectStatusActive {
			return apperrors.Wrap(apperrors.ErrProjectEnvInactive).
				WithParam("Name", app.ProjectEnvID)
		}
	}

	return nil
}

func (s *service) loadReferenceProjects(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireActive bool,
	errorIfUnavail bool,
	projectIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadProjectIDs := make([]string, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		if _, ok := refObjects.RefProjects[projectID]; ok {
			continue
		}
		loadProjectIDs = append(loadProjectIDs, projectID)
	}
	if len(loadProjectIDs) == 0 {
		return nil
	}

	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("project.status = ?", base.ProjectStatusActive))
	}

	projects, err := s.projectRepo.ListByIDs(ctx, db, loadProjectIDs, opts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, project := range projects {
		refObjects.RefProjects[project.ID] = project
	}

	for _, id := range loadProjectIDs {
		project := refObjects.RefProjects[id]
		if (requireActive || errorIfUnavail) && project == nil {
			return apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", id)
		}
	}

	return nil
}

func (s *service) loadReferenceProjectEnvs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireActive bool,
	errorIfUnavail bool,
	projectEnvIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadProjectEnvIDs := make([]string, 0, len(projectEnvIDs))
	for _, projectEnvID := range projectEnvIDs {
		if projectEnv, ok := refObjects.RefProjectEnvs[projectEnvID]; ok && projectEnv.Project != nil {
			continue
		}
		loadProjectEnvIDs = append(loadProjectEnvIDs, projectEnvID)
	}
	if len(loadProjectEnvIDs) == 0 {
		return nil
	}

	opts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("project_env.status = ?", base.ProjectStatusActive))
	}

	projectEnvs, err := s.projectEnvRepo.ListByIDs(ctx, db, loadProjectEnvIDs, opts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, projectEnv := range projectEnvs {
		refObjects.RefProjectEnvs[projectEnv.ID] = projectEnv
	}

	for _, id := range loadProjectEnvIDs {
		projectEnv := refObjects.RefProjectEnvs[id]
		if (requireActive || errorIfUnavail) && projectEnv == nil {
			return apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", id)
		}
		if projectEnv == nil {
			continue
		}
		if errorIfUnavail && projectEnv.Project == nil {
			return apperrors.Wrap(apperrors.ErrProjectNotFound).
				WithParam("Name", projectEnv.ProjectID)
		}
		if requireActive && (projectEnv.Project == nil || projectEnv.Project.Status != base.ProjectStatusActive) {
			return apperrors.Wrap(apperrors.ErrProjectInactive).
				WithParam("Name", projectEnv.ProjectID)
		}
	}

	return nil
}

func (s *service) loadReferenceUsers(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	errorIfUnavail bool,
	userIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadUserIDs := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if _, ok := refObjects.RefUsers[userID]; ok {
			continue
		}
		loadUserIDs = append(loadUserIDs, userID)
	}
	if len(loadUserIDs) == 0 {
		return nil
	}

	userMap, err := s.userService.LoadUsers(ctx, db, loadUserIDs, errorIfUnavail)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, user := range userMap {
		refObjects.RefUsers[user.ID] = user
	}

	return nil
}
