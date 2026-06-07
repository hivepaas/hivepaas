package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type ProjectRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.Project, error)
	GetByIDAndOwner(ctx context.Context, db database.IDB, projectID, ownerID string,
		opts ...bunex.SelectQueryOption) (*entity.Project, error)
	GetByName(ctx context.Context, db database.IDB, name string,
		opts ...bunex.SelectQueryOption) (*entity.Project, error)
	GetByKey(ctx context.Context, db database.IDB, key string,
		opts ...bunex.SelectQueryOption) (*entity.Project, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Project, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.Project, error)

	Upsert(ctx context.Context, db database.IDB, project *entity.Project,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, projects []*entity.Project,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, project *entity.Project,
		opts ...bunex.UpdateQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type projectRepo struct {
}

func NewProjectRepo() ProjectRepo {
	return &projectRepo{}
}

func (repo *projectRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.Project, error) {
	project := &entity.Project{}
	query := db.NewSelect().Model(project).Where("project.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if project == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Project").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return project, nil
}

func (repo *projectRepo) GetByIDAndOwner(ctx context.Context, db database.IDB, projectID, ownerID string,
	opts ...bunex.SelectQueryOption) (*entity.Project, error) {
	project := &entity.Project{}
	query := db.NewSelect().Model(project).
		Where("project.id = ?", projectID).
		Where("project.owner_id = ?", ownerID)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if project == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Project").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return project, nil
}

func (repo *projectRepo) GetByName(ctx context.Context, db database.IDB, name string,
	opts ...bunex.SelectQueryOption) (*entity.Project, error) {
	project := &entity.Project{}
	query := db.NewSelect().Model(project).Where("LOWER(project.name) = ?", strings.ToLower(name))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if project == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Project").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return project, nil
}

func (repo *projectRepo) GetByKey(ctx context.Context, db database.IDB, key string,
	opts ...bunex.SelectQueryOption) (*entity.Project, error) {
	project := &entity.Project{}
	query := db.NewSelect().Model(project).Where("LOWER(project.key) = ?", strings.ToLower(key)).Limit(1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if project == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Project").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return project, nil
}

func (repo *projectRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Project, *basedto.PagingMeta, error) {
	var projects []*entity.Project
	query := db.NewSelect().Model(&projects)
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}

	return projects, pagingMeta, nil
}

func (repo *projectRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.Project, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var projects []*entity.Project
	query := db.NewSelect().Model(&projects).Where("project.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return projects, nil
}

func (repo *projectRepo) Upsert(ctx context.Context, db database.IDB, project *entity.Project,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Project{project}, conflictCols, updateCols, opts...)
}

func (repo *projectRepo) UpsertMulti(ctx context.Context, db database.IDB, projects []*entity.Project,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(projects) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&projects)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectRepo) Update(ctx context.Context, db database.IDB, project *entity.Project,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(project).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Project)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
