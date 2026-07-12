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

type LoginTrustedDeviceRepo interface {
	GetByUserAndDevice(ctx context.Context, db database.IDB, userID, deviceID string,
		opts ...bunex.SelectQueryOption) (*entity.LoginTrustedDevice, error)

	Upsert(ctx context.Context, db database.IDB, loginTrustedDevice *entity.LoginTrustedDevice,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, loginTrustedDevices []*entity.LoginTrustedDevice,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type loginTrustedDeviceRepo struct {
}

func NewLoginTrustedDeviceRepo() LoginTrustedDeviceRepo {
	return &loginTrustedDeviceRepo{}
}

func (repo *loginTrustedDeviceRepo) GetByUserAndDevice(
	ctx context.Context,
	db database.IDB,
	userID, deviceID string,
	opts ...bunex.SelectQueryOption,
) (*entity.LoginTrustedDevice, error) {
	trustedDevice := &entity.LoginTrustedDevice{}
	query := db.NewSelect().Model(trustedDevice).
		Where("login_trusted_device.user_id = ?", userID).
		Where("login_trusted_device.device_id = ?", deviceID).
		Limit(1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if trustedDevice == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Login trusted device").WithCause(err).
			WithMsgLog("user id: %s, device id: %s", userID, deviceID)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return trustedDevice, nil
}

func (repo *loginTrustedDeviceRepo) Upsert(ctx context.Context, db database.IDB,
	loginTrustedDevice *entity.LoginTrustedDevice,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.LoginTrustedDevice{loginTrustedDevice}, conflictCols, updateCols, opts...)
}

func (repo *loginTrustedDeviceRepo) UpsertMulti(ctx context.Context, db database.IDB,
	loginTrustedDevices []*entity.LoginTrustedDevice,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(loginTrustedDevices) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&loginTrustedDevices)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *loginTrustedDeviceRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.LoginTrustedDevice)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
