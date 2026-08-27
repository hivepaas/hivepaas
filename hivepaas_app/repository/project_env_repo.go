package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type ProjectEnvRepo interface {
	GetByID(ctx context.Context, db database.IDB, projectID, id string,
		opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error)
	GetByName(ctx context.Context, db database.IDB, projectID, name string,
		opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error)
	GetByKey(ctx context.Context, db database.IDB, projectID, key string,
		opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error)
	List(ctx context.Context, db database.IDB, projectID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ProjectEnv, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.ProjectEnv, error)

	Upsert(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, projectEnvs []*entity.ProjectEnv,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
		opts ...bunex.UpdateQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type projectEnvRepo struct {
}

func NewProjectEnvRepo() ProjectEnvRepo {
	return &projectEnvRepo{}
}

func (repo *projectEnvRepo) GetByID(ctx context.Context, db database.IDB, projectID, id string,
	opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error) {
	projectEnv := &entity.ProjectEnv{}
	query := db.NewSelect().Model(projectEnv).Where("project_env.id = ?", id)
	if projectID != "" {
		query = query.Where("project_env.project_id = ?", projectID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if projectEnv == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("ProjectEnv").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return projectEnv, nil
}

func (repo *projectEnvRepo) GetByName(ctx context.Context, db database.IDB, projectID, name string,
	opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error) {
	projectEnv := &entity.ProjectEnv{}
	query := db.NewSelect().Model(projectEnv).
		Where("project_env.project_id = ?", projectID).
		Where("project_env.name = ?", name)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if projectEnv == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("ProjectEnv").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return projectEnv, nil
}

func (repo *projectEnvRepo) GetByKey(ctx context.Context, db database.IDB, projectID, key string,
	opts ...bunex.SelectQueryOption) (*entity.ProjectEnv, error) {
	projectEnv := &entity.ProjectEnv{}
	query := db.NewSelect().Model(projectEnv).
		Where("project_env.project_id = ?", projectID).
		Where("project_env.key = ?", key).
		Limit(1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if projectEnv == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("ProjectEnv").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return projectEnv, nil
}

func (repo *projectEnvRepo) List(ctx context.Context, db database.IDB, projectID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ProjectEnv, *basedto.PagingMeta, error) {
	var projectEnvs []*entity.ProjectEnv
	query := db.NewSelect().Model(&projectEnvs)
	if projectID != "" {
		query = query.Where("project_env.project_id = ?", projectID)
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

	return projectEnvs, pagingMeta, nil
}

func (repo *projectEnvRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.ProjectEnv, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var projectEnvs []*entity.ProjectEnv
	query := db.NewSelect().Model(&projectEnvs).Where("project_env.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return projectEnvs, nil
}

func (repo *projectEnvRepo) Upsert(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.ProjectEnv{projectEnv}, conflictCols, updateCols, opts...)
}

func (repo *projectEnvRepo) UpsertMulti(ctx context.Context, db database.IDB, projectEnvs []*entity.ProjectEnv,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(projectEnvs) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&projectEnvs)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *projectEnvRepo) Update(ctx context.Context, db database.IDB, projectEnv *entity.ProjectEnv,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(projectEnv).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *projectEnvRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.ProjectEnv)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
