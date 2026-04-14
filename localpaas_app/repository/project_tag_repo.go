package repository

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type ProjectTagRepo interface {
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ProjectTag, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, projectTags []*entity.ProjectTag,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error

	DeleteAllByProjects(ctx context.Context, db database.IDB, projectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type projectTagRepo struct {
}

func NewProjectTagRepo() ProjectTagRepo {
	return &projectTagRepo{}
}

func (repo *projectTagRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ProjectTag, *basedto.PagingMeta, error) {
	var projectTags []*entity.ProjectTag
	query := db.NewSelect().Model(&projectTags)
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

	return projectTags, pagingMeta, nil
}

func (repo *projectTagRepo) UpsertMulti(ctx context.Context, db database.IDB, projectTags []*entity.ProjectTag,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(projectTags) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&projectTags)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectTagRepo) DeleteAllByProjects(ctx context.Context, db database.IDB,
	projectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(projectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.ProjectTag)(nil)).
		Where("project_id IN (?)", bun.List(projectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectTagRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.ProjectTag)(nil)).ForceDelete()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
