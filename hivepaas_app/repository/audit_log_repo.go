package repository

import (
	"context"
	"database/sql"
	"errors"

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
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.AuditLog, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, auditLog *entity.AuditLog,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, auditLogs []*entity.AuditLog,
		opts ...bunex.InsertQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type auditLogRepo struct {
}

func NewAuditLogRepo() AuditLogRepo {
	return &auditLogRepo{}
}

func (repo *auditLogRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.AuditLog, error) {
	auditLog := &entity.AuditLog{}
	query := db.NewSelect().Model(auditLog).Where("audit_log.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if auditLog == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("AuditLog").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return auditLog, nil
}

func (repo *auditLogRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error) {
	var auditLogs []*entity.AuditLog
	query := db.NewSelect().Model(&auditLogs)
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
	return auditLogs, pagingMeta, nil
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
