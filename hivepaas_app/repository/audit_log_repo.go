package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// AuditLogRepo is deliberately narrower than the other repositories: an audit
// row can be written and read, and removed only by the retention sweep. There is
// no update and no delete-by-id, because a record that can be edited or picked
// off one row at a time is not evidence of anything.
type AuditLogRepo interface {
	// GetByID takes the same scope as List, and hides the row the same way.
	//
	// A nil scope means no scoping - every row is visible. Passing one is what
	// stops an id from being a way around the boundary the listing enforces: the
	// entries record secret reveals, refused attempts and client addresses, so
	// being able to read one by id from outside its scope would make the scoping
	// decorative.
	GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope, id string,
		opts ...bunex.SelectQueryOption) (*entity.AuditLog, error)
	List(ctx context.Context, db database.IDB, scope *entity.ObjectScope, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, auditLog *entity.AuditLog,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, auditLogs []*entity.AuditLog,
		opts ...bunex.InsertQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type auditLogRepo struct {
	appRepo AppRepo
}

func NewAuditLogRepo(appRepo AppRepo) AuditLogRepo {
	return &auditLogRepo{
		appRepo: appRepo,
	}
}

func (repo *auditLogRepo) GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	id string, opts ...bunex.SelectQueryOption) (*entity.AuditLog, error) {
	theOpts, err := repo.applyScopeFilter(ctx, db, opts, scope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	auditLog := &entity.AuditLog{}
	query := db.NewSelect().Model(auditLog).Where("audit_log.id = ?", id)
	query = bunex.ApplySelect(query, theOpts...)

	err = query.Scan(ctx)
	if auditLog == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("AuditLog").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return auditLog, nil
}

func (repo *auditLogRepo) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error) {
	theOpts, err := repo.applyScopeFilter(ctx, db, opts, scope)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	var auditLogs []*entity.AuditLog
	query := db.NewSelect().Model(&auditLogs)
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
	return auditLogs, pagingMeta, nil
}

// applyScopeFilter narrows a query to what the given scope may see.
//
// One implementation for both reads. Two would be two chances to disagree about
// what a scope can see, and the half that was more generous would be the one
// somebody found.
func (repo *auditLogRepo) applyScopeFilter(ctx context.Context, db database.IDB,
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

func (repo *auditLogRepo) loadScopeData(ctx context.Context, db database.IDB, scope *entity.ObjectScope) error {
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

// applyAppFilter filters settings belong to the app or to any parent object
func (repo *auditLogRepo) applyAppFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("audit_log.object_id = ?", scope.AppID),
		)
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This app's log
			bunex.SelectWhere("audit_log.object_id = ?", scope.AppID),
			// Logs from child apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS child_app ON child_app.id = audit_log.object_id AND "+
					"child_app.deleted_at IS NULL"),
				bunex.SelectWhere("child_app.parent_id = ?", scope.AppID),
			),
		),
	)
}

// applyProjectEnvFilter filters settings belong to the project env or to any parent object
func (repo *auditLogRepo) applyProjectEnvFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("audit_log.object_id = ?", scope.ProjectEnvID))
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This env's log
			bunex.SelectWhere("audit_log.object_id = ?", scope.ProjectEnvID),
			// Logs from containing apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS app ON app.id = audit_log.object_id AND "+
					"app.deleted_at IS NULL"),
				bunex.SelectWhere("app.project_env_id = ?", scope.ProjectEnvID),
			),
		),
	)
}

// applyProjectFilter filters settings belong to the project or to any parent object
func (repo *auditLogRepo) applyProjectFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	projectID := scope.ProjectID

	if scope.NoInherited {
		return append(opts,
			bunex.SelectWhere("audit_log.object_id = ?", projectID))
	}

	return append(opts,
		bunex.SelectWhereGroup(
			// This env's log
			bunex.SelectWhere("audit_log.object_id = ?", projectID),
			// Logs from containing envs
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN project_envs AS env ON env.id = audit_log.object_id AND "+
					"env.deleted_at IS NULL"),
				bunex.SelectWhere("env.project_id = ?", projectID),
			),
			// Logs from containing apps
			bunex.SelectWhereOrGroup(
				bunex.SelectJoin("LEFT JOIN apps AS app ON app.id = audit_log.object_id AND "+
					"app.deleted_at IS NULL"),
				bunex.SelectWhere("app.project_id = ?", projectID),
			),
		),
	)
}

// applyDirectUserFilter filters settings belong to the user
func (repo *auditLogRepo) applyDirectUserFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("audit_log.object_id = ?", scope.UserID))
	return opts
}

// applyGlobalFilter filters settings belong to global scope
func (repo *auditLogRepo) applyGlobalFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts, bunex.SelectWhere("audit_log.object_id IS NULL"))
	}
	return opts
}

// applyHivepaasFilter filters settings belong to Hivepaas scope
func (repo *auditLogRepo) applyHivepaasFilter(opts []bunex.SelectQueryOption) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("audit_log.scope = ?", base.ObjectScopeHivepaas))
	return opts
}

func (repo *auditLogRepo) Insert(ctx context.Context, db database.IDB, auditLog *entity.AuditLog,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.AuditLog{auditLog}, opts...)
}

func (repo *auditLogRepo) InsertMulti(ctx context.Context, db database.IDB, auditLogs []*entity.AuditLog,
	opts ...bunex.InsertQueryOption) error {
	if len(auditLogs) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&auditLogs)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *auditLogRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.AuditLog)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
