package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
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
		return nil, hperrors.NewNotFound("TaskLog").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
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
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = int(total)

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
	_, err := repo.insertMultiQuery(db, logs, opts...).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// insertMultiQuery replaces every NUL byte in the logs' data before they are written.
//
// The data is a process's output, and nothing stops a process from printing one: a
// binary file, /proc/<pid>/cmdline, find -print0. A Postgres text column cannot hold
// a NUL, and bun refuses to send one rather than drop it, which fails the whole batch
// and the transaction it runs in - for a deployment, the one that records how the
// deployment ended.
func (repo *taskLogRepo) insertMultiQuery(db database.IDB, logs []*entity.TaskLog,
	opts ...bunex.InsertQueryOption) *bun.InsertQuery {
	for _, log := range logs {
		log.Data = strings.ReplaceAll(log.Data, "\x00", "\uFFFD")
	}
	query := db.NewInsert().Model(&logs)
	return bunex.ApplyInsert(query, opts...)
}

func (repo *taskLogRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.TaskLog)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
