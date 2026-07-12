package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type SystemStatusRepo interface {
	Get(ctx context.Context, db database.IDB, opts ...bunex.SelectQueryOption) (*entity.SystemStatus, error)

	Upsert(ctx context.Context, db database.IDB, systemStatus *entity.SystemStatus, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
}

type systemStatusRepo struct {
}

func NewSystemStatusRepo() SystemStatusRepo {
	return &systemStatusRepo{}
}

func (repo *systemStatusRepo) Get(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) (*entity.SystemStatus, error) {
	systemStatus := &entity.SystemStatus{}
	query := db.NewSelect().Model(systemStatus).Where("system_status.id = ?", 1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if systemStatus == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("SystemStatus").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return systemStatus, nil
}

func (repo *systemStatusRepo) Upsert(ctx context.Context, db database.IDB, systemStatus *entity.SystemStatus,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	query := db.NewInsert().Model(systemStatus)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
