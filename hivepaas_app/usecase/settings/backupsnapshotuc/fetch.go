package backupsnapshotuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// reachOf is a view's reach for a viewer acting with action, and the
// repositories it takes in.
func (uc *UC) reachOf(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	action base.ActionType,
) (*snapshotReach, []*entity.Setting, error) {
	repos, err := uc.reposOf(ctx, db, scope)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	apps, err := uc.appsOf(ctx, db, scope)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	visibility := uc.PermissionManager.NewVisibility(db, auth)
	var visibilityErr error
	allows := func(projectID, env string) bool {
		allowed, err := visibility.AllowsProjectEnv(ctx, projectID, env, action)
		if err != nil && visibilityErr == nil {
			visibilityErr = err
		}
		return allowed
	}
	reach := computeReach(scope, repos, apps, allows)
	if visibilityErr != nil {
		return nil, nil, hperrors.Wrap(visibilityErr)
	}
	return reach, repos, nil
}

// reposOf are the repositories a view takes in: those its scope sees, and for a
// project the repositories of its envs too.
func (uc *UC) reposOf(ctx context.Context, db database.IDB, scope *entity.ObjectScope) ([]*entity.Setting, error) {
	isRepo := bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupRepo)
	repos, _, err := uc.SettingRepo.List(ctx, db, scope, nil, isRepo)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if scope.ScopeType != base.ObjectScopeProject {
		return repos, nil
	}
	envRepos, _, err := uc.SettingRepo.List(ctx, db, nil, nil, isRepo,
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeProjectEnv),
		bunex.SelectWhere("setting.object_id LIKE ?", scope.ProjectID+":%"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return append(repos, envRepos...), nil
}

// appsOf are the apps whose snapshots a view shows.
func (uc *UC) appsOf(ctx context.Context, db database.IDB, scope *entity.ObjectScope) ([]*entity.App, error) {
	var opts []bunex.SelectQueryOption
	projectID := ""
	switch scope.ScopeType {
	case base.ObjectScopeGlobal:
	case base.ObjectScopeProject:
		projectID = scope.ProjectID
	case base.ObjectScopeProjectEnv:
		projectID = scope.ProjectID
		opts = append(opts, bunex.SelectWhere("app.project_env_id = ?", scope.ProjectEnvID))
	case base.ObjectScopeApp:
		projectID = scope.ProjectID
		opts = append(opts, bunex.SelectWhere("app.id = ?", scope.AppID))
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
		// No view of snapshots lives there.
		return nil, nil
	}
	apps, _, err := uc.appRepo.List(ctx, db, projectID, nil,
		append(opts, bunex.SelectColumns("app.id", "app.name", "app.project_id", "app.project_env_id"))...)
	return apps, hperrors.Wrap(err)
}

// snapshotRefs names what the snapshots' tags point at: their repositories,
// apps and jobs, deleted ones included, so that they read as deleted.
func (uc *UC) snapshotRefs(
	ctx context.Context,
	db database.IDB,
	repos []*entity.Setting,
	tagsByRecord map[string][]string,
) (*backupsnapshotdto.SnapshotRefs, error) {
	refs := &backupsnapshotdto.SnapshotRefs{
		Repos: make(map[string]*entity.Setting, len(repos)),
		Apps:  map[string]*entity.App{},
		Jobs:  map[string]*entity.Setting{},
	}
	for _, repo := range repos {
		refs.Repos[repo.ID] = repo
	}
	var appIDs, jobIDs []string
	for _, tags := range tagsByRecord {
		parsed := entity.ParseDataBackupSnapshotTags(tags)
		if parsed.AppID != "" {
			appIDs = append(appIDs, parsed.AppID)
		}
		if parsed.JobID != "" {
			jobIDs = append(jobIDs, parsed.JobID)
		}
	}
	if len(appIDs) > 0 {
		apps, err := uc.appRepo.ListByIDs(ctx, db, "", appIDs, bunex.SelectWithDeleted())
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, app := range apps {
			refs.Apps[app.ID] = app
		}
	}
	if len(jobIDs) > 0 {
		jobs, _, err := uc.SettingRepo.List(ctx, db, nil, nil, bunex.SelectWithDeleted(),
			bunex.SelectWhere("setting.id IN (?)", bunex.List(jobIDs)))
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, job := range jobs {
			refs.Jobs[job.ID] = job
		}
	}
	return refs, nil
}

// tagsOf are the records' tags, by record, in their order.
func (uc *UC) tagsOf(ctx context.Context, db database.IDB, records []*entity.Setting) (map[string][]string, error) {
	byRecord := make(map[string][]string, len(records))
	if len(records) == 0 {
		return byRecord, nil
	}
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	tags, _, err := uc.TagRepo.List(ctx, db, nil,
		bunex.SelectWhere("tag.object_id IN (?)", bunex.List(ids)),
		bunex.SelectOrder("tag.object_id", "tag.index"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, tag := range tags {
		byRecord[tag.ObjectID] = append(byRecord[tag.ObjectID], tag.Tag)
	}
	return byRecord, nil
}
