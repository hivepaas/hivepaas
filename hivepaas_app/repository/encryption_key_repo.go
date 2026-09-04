package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type EncryptionKeyRepo interface {
	GetActive(ctx context.Context, db database.IDB,
		opts ...bunex.SelectQueryOption) (*entity.EncryptionKey, error)

	Insert(ctx context.Context, db database.IDB, key *entity.EncryptionKey,
		opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, key *entity.EncryptionKey,
		opts ...bunex.UpdateQueryOption) error
}

type encryptionKeyRepo struct {
}

func NewEncryptionKeyRepo() EncryptionKeyRepo {
	return &encryptionKeyRepo{}
}

// GetActive returns the key in use, or nil when the app has never created one.
func (repo *encryptionKeyRepo) GetActive(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) (*entity.EncryptionKey, error) {
	key := &entity.EncryptionKey{}
	query := db.NewSelect().Model(key).Where("is_active = ?", true)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return key, nil
}

func (repo *encryptionKeyRepo) Insert(ctx context.Context, db database.IDB,
	key *entity.EncryptionKey, opts ...bunex.InsertQueryOption) error {
	query := db.NewInsert().Model(key)
	query = bunex.ApplyInsert(query, opts...)

	if _, err := query.Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *encryptionKeyRepo) Update(ctx context.Context, db database.IDB,
	key *entity.EncryptionKey, opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(key).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	if _, err := query.Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
