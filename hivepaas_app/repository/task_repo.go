package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type TaskRepo interface {
	GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.TaskType, id string,
		opts ...bunex.SelectQueryOption) (*entity.Task, error)
	ListByTarget(ctx context.Context, db database.IDB, targetID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.Task, error)

	Insert(ctx context.Context, db database.IDB, task *entity.Task,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, task *entity.Task, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, task *entity.Task,
		opts ...bunex.UpdateQueryOption) error

	DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteAllByApps(ctx context.Context, db database.IDB, appIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteAllByProjects(ctx context.Context, db database.IDB, projectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteAllByProjectEnvs(ctx context.Context, db database.IDB, projectEnvIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteAllByUsers(ctx context.Context, db database.IDB, userIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB,
		opts ...bunex.DeleteQueryOption) error
}

type taskRepo struct {
	appRepo AppRepo
}

func NewTaskRepo(appRepo AppRepo) TaskRepo {
	return &taskRepo{
		appRepo: appRepo,
	}
}

func (repo *taskRepo) GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.TaskType, id string, opts ...bunex.SelectQueryOption) (*entity.Task, error) {
	theOpts, err := repo.applyScopeFilter(ctx, db, opts, scope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	task := &entity.Task{}
	query := db.NewSelect().Model(task).Where("task.id = ?", id)
	if typ != "" {
		query = query.Where("task.type = ?", typ)
	}
	query = bunex.ApplySelect(query, theOpts...)

	err = query.Scan(ctx)
	if task == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("Task").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

func (repo *taskRepo) ListByTarget(ctx context.Context, db database.IDB, targetID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks)
	if targetID != "" {
		query = query.Where("task.target_id = ?", targetID)
	}
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}
	return tasks, pagingMeta, nil
}

func (repo *taskRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.Task, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks).Where("task.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return tasks, nil
}

func (repo *taskRepo) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	theOpts, err := repo.applyScopeFilter(ctx, db, opts, scope)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks)
	query = bunex.ApplySelect(query, theOpts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	if err = query.Scan(ctx); err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}
	return tasks, pagingMeta, nil
}

// applyScopeFilter narrows a query to what the given scope may see.
//
// One implementation for both reads. Two would be two chances to disagree about
// what a scope can see, and the half that was more generous would be the one
// somebody found.
func (repo *taskRepo) applyScopeFilter(ctx context.Context, db database.IDB,
	opts []bunex.SelectQueryOption, scope *entity.ObjectScope) ([]bunex.SelectQueryOption, error) {
	if scope == nil {
		return opts, nil
	}
	if err := repo.loadScopeData(ctx, db, scope); err != nil {
		return nil, hperrors.Wrap(err)
	}

	switch scope.ScopeType {
	case base.ObjectScopeApp:
		return repo.applyAppFilter(opts, scope), nil
	case base.ObjectScopeProjectEnv:
		return repo.applyProjectEnvFilter(opts, scope), nil
	case base.ObjectScopeProject:
		return repo.applyProjectFilter(opts, scope), nil
	case base.ObjectScopeUser:
		return repo.applyDirectUserFilter(opts, scope), nil
	case base.ObjectScopeGlobal:
		return repo.applyGlobalFilter(opts, scope), nil
	case base.ObjectScopeHivepaas:
		return repo.applyHivepaasFilter(opts), nil
	default:
		return opts, nil
	}
}

func (repo *taskRepo) loadScopeData(ctx context.Context, db database.IDB, scope *entity.ObjectScope) error {
	if scope == nil {
		return nil
	}

	if scope.ScopeType == base.ObjectScopeApp && (scope.ProjectID == "" || scope.ProjectEnvID == "") {
		app, err := repo.appRepo.GetByID(ctx, db, "", scope.AppID,
			bunex.SelectColumns("project_id", "project_env_id", "parent_id"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		scope.ProjectID = app.ProjectID
		scope.ProjectEnvID = app.ProjectEnvID
		scope.ParentAppID = app.ParentID
	}

	return nil
}

// applyAppFilter filters tasks belong to the app or to any parent object
func (repo *taskRepo) applyAppFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("task.object_id = ?", scope.AppID),
		)
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This app's task
			bunex.SelectWhere("task.object_id = ?", scope.AppID),
			// Tasks from child apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS child_app ON child_app.id = task.object_id AND "+
					"child_app.deleted_at IS NULL"),
				bunex.SelectWhere("child_app.parent_id = ?", scope.AppID),
			),
		),
	)
}

// applyProjectEnvFilter filters tasks belong to the project env or to any parent object
func (repo *taskRepo) applyProjectEnvFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("task.object_id = ?", scope.ProjectEnvID))
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This env's task
			bunex.SelectWhere("task.object_id = ?", scope.ProjectEnvID),
			// Tasks from containing apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS app ON app.id = task.object_id AND "+
					"app.deleted_at IS NULL"),
				bunex.SelectWhere("app.project_env_id = ?", scope.ProjectEnvID),
			),
		),
	)
}

// applyProjectFilter filters tasks belong to the project or to any parent object
func (repo *taskRepo) applyProjectFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	projectID := scope.ProjectID

	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("task.object_id = ?", projectID))
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This env's tasks
			bunex.SelectWhere("task.object_id = ?", projectID),
			// Tasks from containing envs
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN project_envs AS env ON env.id = task.object_id AND "+
					"env.deleted_at IS NULL"),
				bunex.SelectWhere("env.project_id = ?", projectID),
			),
			// Tasks from containing apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS app ON app.id = task.object_id AND "+
					"app.deleted_at IS NULL"),
				bunex.SelectWhere("app.project_id = ?", projectID),
			),
		),
	)
}

// applyDirectUserFilter filters tasks belong to the user
func (repo *taskRepo) applyDirectUserFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("task.object_id = ?", scope.UserID))
	return opts
}

// applyGlobalFilter filters tasks belong to global scope
func (repo *taskRepo) applyGlobalFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts, bunex.SelectWhere("task.object_id IS NULL"))
	}
	return opts
}

// applyHivepaasFilter filters tasks belong to Hivepaas scope
func (repo *taskRepo) applyHivepaasFilter(opts []bunex.SelectQueryOption) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("task.scope = ?", base.ObjectScopeHivepaas))
	return opts
}

func (repo *taskRepo) Insert(ctx context.Context, db database.IDB, task *entity.Task,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.Task{task}, opts...)
}

func (repo *taskRepo) InsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	opts ...bunex.InsertQueryOption) error {
	if len(tasks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&tasks)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) Upsert(ctx context.Context, db database.IDB, task *entity.Task,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Task{task}, conflictCols, updateCols, opts...)
}

func (repo *taskRepo) UpsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(tasks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&tasks)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) Update(ctx context.Context, db database.IDB, task *entity.Task,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(task).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// NOTE: this UpdateMulti may not work properly with JSON columns defined as string in the entity struct
func (repo *taskRepo) UpdateMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(&tasks).Bulk()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(ids) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.id IN (?)", bun.List(ids))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteAllByApps(ctx context.Context, db database.IDB, appIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(appIDs) == 0 {
		return nil
	}
	subQuery := db.NewSelect().Model((*entity.Task)(nil)).
		Column("task.id").
		Where("(task.target_id IN (?) OR EXISTS(SELECT 1 FROM deployments WHERE "+
			"deployments.id = task.target_id AND deployments.app_id IN (?)))", bun.List(appIDs), bun.List(appIDs)).
		For("UPDATE SKIP LOCKED")

	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.id IN (?)", subQuery)
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteAllByProjects(ctx context.Context, db database.IDB, projectIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(projectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.target_id IN (?)", bun.List(projectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteAllByProjectEnvs(ctx context.Context, db database.IDB, projectEnvIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(projectEnvIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.target_id IN (?)", bun.List(projectEnvIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteAllByUsers(ctx context.Context, db database.IDB, userIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.target_id IN (?)", bun.List(userIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
