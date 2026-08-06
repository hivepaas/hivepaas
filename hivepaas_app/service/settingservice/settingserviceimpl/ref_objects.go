package settingserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

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
	inSettings ...*entity.Setting,
) error {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.loadRefObjectsByIDs(ctx, db, refObjects, scope, true, requireActive, allRefIDs)
}

func (s *service) LoadRefObjectsSkipMissing(
	ctx context.Context,
	db database.IDB,
	refObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	inSettings ...*entity.Setting,
) (err error) {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.loadRefObjectsByIDs(ctx, db, refObjects, scope, false, requireActive, allRefIDs)
}

func (s *service) LoadRefObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
) error {
	return s.loadRefObjectsByIDs(ctx, db, pRefObjects, scope, true, requireActive, refIDs)
}

func (s *service) LoadRefObjectsByIDsSkipMissing(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
) error {
	return s.loadRefObjectsByIDs(ctx, db, pRefObjects, scope, false, requireActive, refIDs)
}

func (s *service) loadRefObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireExistence bool,
	requireActive bool,
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
		err = s.loadReferenceUsers(ctx, db, pRefObjects, requireExistence, requireActive, refIDs.RefUserIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref apps
	if len(refIDs.RefAppIDs) > 0 {
		err = s.loadReferenceApps(ctx, db, pRefObjects, requireExistence, requireActive, refIDs.RefAppIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref projects
	if len(refIDs.RefProjectIDs) > 0 {
		err = s.loadReferenceProjects(ctx, db, pRefObjects, requireExistence, requireActive, refIDs.RefProjectIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref project envs
	if len(refIDs.RefProjectEnvIDs) > 0 {
		err = s.loadReferenceProjectEnvs(ctx, db, pRefObjects, requireExistence, requireActive, refIDs.RefProjectEnvIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Load ref settings
	if len(refIDs.RefSettingIDs) > 0 {
		err = s.loadReferenceSettings(ctx, db, pRefObjects, scope, requireExistence, requireActive, refIDs.RefSettingIDs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Calculate recursive ref IDs to load
	newRecursiveRefIDs := refIDs.GetRecursiveRefObjectIDs(refObjects)
	if !newRecursiveRefIDs.HasData() {
		return nil
	}

	err = s.loadRefObjectsByIDs(ctx, db, pRefObjects, scope, requireExistence, requireActive, newRecursiveRefIDs)
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
	requireExistence bool,
	requireActive bool,
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

	for _, id := range loadSettingIDs {
		setting := refObjects.RefSettings[id]
		if setting == nil {
			if requireExistence {
				return apperrors.Wrap(apperrors.ErrSettingNotFound).WithParam("Name", id)
			}
			continue
		}
	}

	return nil
}

func (s *service) loadReferenceApps(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireExistence bool,
	requireActive bool,
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
		if app == nil {
			if requireExistence {
				return apperrors.Wrap(apperrors.ErrAppNotFound).WithParam("Name", id)
			}
			continue
		}
		if requireActive {
			if app.Status != base.AppStatusActive {
				return apperrors.Wrap(apperrors.ErrAppInactive).WithParam("Name", app.Name)
			}
			if app.Project == nil || app.Project.Status != base.ProjectStatusActive {
				return apperrors.Wrap(apperrors.ErrProjectInactive).WithParam("Name", app.ProjectID)
			}
			if app.ProjectEnv == nil || app.ProjectEnv.Status != base.ProjectStatusActive {
				return apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithParam("Name", app.ProjectEnvID)
			}
		}
	}

	return nil
}

func (s *service) loadReferenceProjects(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireExistence bool,
	requireActive bool,
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
		if project == nil && requireExistence {
			return apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", id)
		}
	}
	return nil
}

func (s *service) loadReferenceProjectEnvs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireExistence bool,
	requireActive bool,
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
		if projectEnv == nil {
			if requireExistence {
				return apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", id)
			}
			continue
		}
		if requireActive {
			if projectEnv.Status != base.ProjectStatusActive {
				return apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithParam("Name", projectEnv.Name)
			}
			if projectEnv.Project == nil || projectEnv.Project.Status != base.ProjectStatusActive {
				return apperrors.Wrap(apperrors.ErrProjectInactive).WithParam("Name", projectEnv.ProjectID)
			}
		}
	}
	return nil
}

func (s *service) loadReferenceUsers(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireExistence bool,
	errorIfUnavailable bool,
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

	loadUsersFunc := gofn.If(requireExistence, s.userService.LoadUsers, s.userService.LoadUsersSkipMissing)
	userMap, err := loadUsersFunc(ctx, db, loadUserIDs, errorIfUnavailable)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, user := range userMap {
		refObjects.RefUsers[user.ID] = user
	}

	return nil
}
