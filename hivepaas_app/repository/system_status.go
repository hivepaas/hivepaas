package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type SystemStatusRepo interface {
	Get(ctx context.Context, db database.IDB, opts ...bunex.SelectQueryOption) (*entity.SystemStatus, error)

	Upsert(ctx context.Context, db database.IDB, systemStatus *entity.SystemStatus, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error

	// RecordLicenseSeenAt raises the latest time license checks have seen to t,
	// and returns that time afterwards: t, or a later one recorded before. It
	// never lowers it, which is the point: a clock set back cannot take back
	// what was seen.
	RecordLicenseSeenAt(ctx context.Context, db database.IDB, t time.Time) (time.Time, error)
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
		return nil, hperrors.NewNotFound("SystemStatus").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *systemStatusRepo) RecordLicenseSeenAt(ctx context.Context, db database.IDB,
	t time.Time) (time.Time, error) {
	var seenAt time.Time
	err := repo.recordLicenseSeenAtQuery(db, t).Scan(ctx, &seenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, hperrors.NewNotFound("SystemStatus").WithCause(err)
	}
	if err != nil {
		return time.Time{}, hperrors.Wrap(err)
	}
	return seenAt.UTC(), nil
}

// recordLicenseSeenAtQuery leaves update_ver and updated_at alone: the time
// moves on every hour or so, and nothing that reads the row acts on that.
// GREATEST ignores a NULL, so the first time recorded is taken as it is.
func (repo *systemStatusRepo) recordLicenseSeenAtQuery(db bun.IDB, t time.Time) *bun.UpdateQuery {
	return db.NewUpdate().Model((*entity.SystemStatus)(nil)).
		Set("license_seen_at = GREATEST(license_seen_at, ?)", t).
		Where("system_status.id = ?", 1).
		Returning("license_seen_at")
}
