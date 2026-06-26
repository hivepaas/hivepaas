package webhookuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (uc *UC) findAppsMatchingRepository(
	ctx context.Context,
	db database.IDB,
	repoID, repoRef string,
	extraAppOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	// Finds all deployment settings which are linked to the repo ID (URL)
	settingListOpts := []bunex.SelectQueryOption{
		bunex.SelectColumns("id", "type", "scope", "object_id"),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'activeMethod' = ?", base.DeploymentMethodRepo),
		bunex.SelectJoin("JOIN res_links ON res_links.src_id = setting.id"),
		bunex.SelectWhere("res_links.deleted_at IS NULL"),
		bunex.SelectWhere("res_links.dst_type = ?", base.ResourceTypeRepo),
		bunex.SelectWhere("res_links.dst_id = ?", repoID),
	}
	if repoRef != "" {
		settingListOpts = append(settingListOpts,
			bunex.SelectWhere("setting.data->'repoSource'->>'repoRef' = ?", repoRef),
		)
	}

	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil, settingListOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}

	appIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		appIDs = append(appIDs, setting.ObjectID)
	}

	appListOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectWhereIn("app.id IN (?)", appIDs...),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectWhere("project.status = ?", base.ProjectStatusActive),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	}
	appListOpts = append(appListOpts, extraAppOpts...)

	apps, _, err := uc.appRepo.List(ctx, db, "", nil, appListOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(apps) == 0 {
		return nil, nil
	}

	matchingApps := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		if app.Project == nil || app.Project.Status != base.ProjectStatusActive {
			continue
		}
		deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
		if deploymentSetting == nil {
			continue
		}
		deploymentSettings := deploymentSetting.MustAsAppDeploymentSettings()
		if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo ||
			deploymentSettings.RepoSource == nil ||
			deploymentSettings.RepoSource.RepoID != repoID ||
			(repoRef != "" && deploymentSettings.RepoSource.RepoRef != repoRef) {
			continue
		}
		matchingApps = append(matchingApps, app)
	}
	return matchingApps, nil
}
