package appserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) FindAppsMatchingRepository(
	ctx context.Context,
	db database.IDB,
	repoID, repoRef string,
	extraAppOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	settingOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.data->>'activeMethod' = ?", base.DeploymentMethodRepo),
	}
	if repoRef != "" {
		settingOpts = append(settingOpts,
			bunex.SelectWhere("setting.data->'repoSource'->>'repoRef' = ?", repoRef),
		)
	}
	matches := func(settings *entity.AppDeploymentSettings) bool {
		return settings.ActiveMethod == base.DeploymentMethodRepo &&
			settings.RepoSource != nil &&
			settings.RepoSource.RepoID == repoID &&
			(repoRef == "" || settings.RepoSource.RepoRef == repoRef)
	}
	return s.findAppsLinkedToRepo(ctx, db, repoID, settingOpts, matches, extraAppOpts...)
}

func (s *service) FindAppsDeployingOnPush(
	ctx context.Context,
	db database.IDB,
	repoID, repoRef string,
	extraAppOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	settingOpts := []bunex.SelectQueryOption{
		bunex.SelectWhereIn("setting.data->>'activeMethod' IN (?)",
			base.DeploymentMethodRepo, base.DeploymentMethodFunction),
	}
	matches := func(settings *entity.AppDeploymentSettings) bool {
		return settings.DeploysOnPush(repoID, repoRef)
	}
	return s.findAppsLinkedToRepo(ctx, db, repoID, settingOpts, matches, extraAppOpts...)
}

// findAppsLinkedToRepo finds the active apps whose active deployment settings
// are linked to the repository, narrowed by settingOpts, and keeps those whose
// settings match.
func (s *service) findAppsLinkedToRepo(
	ctx context.Context,
	db database.IDB,
	repoID string,
	settingOpts []bunex.SelectQueryOption,
	matches func(*entity.AppDeploymentSettings) bool,
	extraAppOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	// Finds all deployment settings which are linked to the repo ID (URL)
	settingListOpts := []bunex.SelectQueryOption{
		bunex.SelectColumns("id", "type", "scope", "object_id"),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectJoin("JOIN res_links ON res_links.src_id = setting.id"),
		bunex.SelectWhere("res_links.deleted_at IS NULL"),
		bunex.SelectWhere("res_links.dst_type = ?", base.ResourceTypeRepo),
		bunex.SelectWhere("res_links.dst_id = ?", repoID),
	}
	settingListOpts = append(settingListOpts, settingOpts...)

	settings, _, err := s.settingRepo.List(ctx, db, nil, nil, settingListOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}

	appIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		appIDs = append(appIDs, setting.ObjectID)
	}

	appListOpts := []bunex.SelectQueryOption{
		bunex.SelectWhereIn("app.id IN (?)", appIDs...),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectWhere("project.status = ?", base.ProjectStatusActive),
		),
		bunex.SelectRelation("ProjectEnv",
			bunex.SelectWhere("project_env.status = ?", base.ProjectStatusActive),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	}
	appListOpts = append(appListOpts, extraAppOpts...)

	apps, _, err := s.appRepo.List(ctx, db, "", nil, appListOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, nil
	}

	matchingApps := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		if app.Project == nil || app.Project.Status != base.ProjectStatusActive {
			continue
		}
		if app.ProjectEnv == nil || app.ProjectEnv.Status != base.ProjectStatusActive {
			continue
		}
		deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
		if deploymentSetting == nil {
			continue
		}
		if !matches(deploymentSetting.MustAsAppDeploymentSettings()) {
			continue
		}
		matchingApps = append(matchingApps, app)
	}
	return matchingApps, nil
}
