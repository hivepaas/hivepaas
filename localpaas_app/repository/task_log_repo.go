package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type TaskLogRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.TaskLog, error)
	List(ctx context.Context, db database.IDB, taskID, targetID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.TaskLog, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, log *entity.TaskLog,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, logs []*entity.TaskLog,
		opts ...bunex.InsertQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type taskLogRepo struct {
}

func NewTaskLogRepo() TaskLogRepo {
	return &taskLogRepo{}
}

func (repo *taskLogRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.TaskLog, error) {
	log := &entity.TaskLog{}
	query := db.NewSelect().Model(log).Where("task_log.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if log == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("TaskLog").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return log, nil
}

func (repo *taskLogRepo) List(ctx context.Context, db database.IDB, taskID, targetID string,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.TaskLog, *basedto.PagingMeta, error) {
	var logs []*entity.TaskLog
	query := db.NewSelect().Model(&logs)
	if taskID != "" {
		query = query.Where("task_log.task_id = ?", taskID)
	}
	if targetID != "" {
		query = query.Where("task_log.target_id = ?", targetID)
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
	return logs, pagingMeta, nil
}

func (repo *taskLogRepo) Insert(ctx context.Context, db database.IDB, log *entity.TaskLog,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.TaskLog{log}, opts...)
}

func (repo *taskLogRepo) InsertMulti(ctx context.Context, db database.IDB, logs []*entity.TaskLog,
	opts ...bunex.InsertQueryOption) error {
	if len(logs) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&logs)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskLogRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.TaskLog)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
