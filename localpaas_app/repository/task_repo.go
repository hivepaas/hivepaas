package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type TaskRepo interface {
	GetByID(ctx context.Context, db database.IDB, typ base.TaskType, id string,
		opts ...bunex.SelectQueryOption) (*entity.Task, error)
	List(ctx context.Context, db database.IDB, jobID string, paging *basedto.Paging,
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

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type taskRepo struct {
}

func NewTaskRepo() TaskRepo {
	return &taskRepo{}
}

func (repo *taskRepo) GetByID(ctx context.Context, db database.IDB, typ base.TaskType, id string,
	opts ...bunex.SelectQueryOption) (*entity.Task, error) {
	task := &entity.Task{}
	query := db.NewSelect().Model(task).Where("task.id = ?", id)
	if typ != "" {
		query = query.Where("task.type = ?", typ)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if task == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Task").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return task, nil
}

func (repo *taskRepo) List(ctx context.Context, db database.IDB, jobID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks)
	if jobID != "" {
		query = query.Where("task.target_id = ?", jobID)
	}
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
		return nil, apperrors.New(err)
	}
	return tasks, nil
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
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) Update(ctx context.Context, db database.IDB, task *entity.Task,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(task).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.Task)(nil)).ForceDelete()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
