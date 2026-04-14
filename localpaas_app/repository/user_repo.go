package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

type UserRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string, opts ...bunex.SelectQueryOption) (
		*entity.User, error)
	GetByUsernameOrEmail(ctx context.Context, db database.IDB, username, email string,
		opts ...bunex.SelectQueryOption) (*entity.User, error)
	GetByUsername(ctx context.Context, db database.IDB, username string,
		opts ...bunex.SelectQueryOption) (*entity.User, error)
	GetByEmail(ctx context.Context, db database.IDB, email string, opts ...bunex.SelectQueryOption) (
		*entity.User, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging, opts ...bunex.SelectQueryOption) (
		[]*entity.User, *basedto.PagingMeta, error)
	ListByEmails(ctx context.Context, db database.IDB, emails []string, opts ...bunex.SelectQueryOption) (
		[]*entity.User, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string, opts ...bunex.SelectQueryOption) (
		[]*entity.User, error)

	Insert(ctx context.Context, db database.IDB, user *entity.User, opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, user *entity.User,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, users []*entity.User,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, user *entity.User, opts ...bunex.UpdateQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type userRepo struct {
}

func NewUserRepo() UserRepo {
	return &userRepo{}
}

func (repo *userRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.User, error) {
	user := &entity.User{}
	query := db.NewSelect().Model(user).Where("\"user\".id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if user == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("User").WithCause(err).WithMsgLog("user id: %s", id)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return user, nil
}

func (repo *userRepo) GetByUsernameOrEmail(ctx context.Context, db database.IDB, username, email string,
	opts ...bunex.SelectQueryOption) (*entity.User, error) {
	user := &entity.User{}
	query := db.NewSelect().Model(user).
		Where("\"user\".username = ?", username).
		WhereOr("\"user\".email = ?", strings.ToLower(email)).
		Limit(1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if user == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("User").WithCause(err).
			WithMsgLog("user name: %s, email: %s", username, email)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return user, nil
}

func (repo *userRepo) GetByUsername(ctx context.Context, db database.IDB, username string,
	opts ...bunex.SelectQueryOption) (*entity.User, error) {
	user := &entity.User{}
	query := db.NewSelect().Model(user).Where("\"user\".username = ?", username)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if user == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("User").WithCause(err).WithMsgLog("user name: %s", username)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return user, nil
}

func (repo *userRepo) GetByEmail(ctx context.Context, db database.IDB, email string,
	opts ...bunex.SelectQueryOption) (*entity.User, error) {
	user := &entity.User{}
	query := db.NewSelect().Model(user).Where("\"user\".email = ?", strings.ToLower(email))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if user == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("User").WithCause(err).WithMsgLog("user email: %s", email)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return user, nil
}

func (repo *userRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.User, *basedto.PagingMeta, error) {
	var users []*entity.User
	query := db.NewSelect().Model(&users)
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

	return users, pagingMeta, nil
}

func (repo *userRepo) ListByEmails(ctx context.Context, db database.IDB, emails []string,
	opts ...bunex.SelectQueryOption) ([]*entity.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}
	lowercaseEmails := gofn.MapSlice(emails, strings.ToLower)
	users := make([]*entity.User, 0, len(emails))
	query := db.NewSelect().Model(&users).Where("\"user\".email IN (?)", bun.List(lowercaseEmails))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("user emails: %v", emails)
	}
	return users, nil
}

func (repo *userRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	users := make([]*entity.User, 0, len(ids))
	query := db.NewSelect().Model(&users).Where("\"user\".id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("user ids: %v", ids)
	}
	return users, nil
}

func (repo *userRepo) Insert(ctx context.Context, db database.IDB, user *entity.User,
	opts ...bunex.InsertQueryOption) error {
	if user.ID == "" {
		newID, err := ulid.NewStringULID()
		if err != nil {
			return apperrors.Wrap(err)
		}
		user.ID = newID
	}

	query := db.NewInsert().Model(user)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *userRepo) Upsert(ctx context.Context, db database.IDB, user *entity.User,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.User{user}, conflictCols, updateCols, opts...)
}

func (repo *userRepo) UpsertMulti(ctx context.Context, db database.IDB, users []*entity.User,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(users) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&users)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *userRepo) Update(ctx context.Context, db database.IDB, user *entity.User,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(user).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *userRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.User)(nil)).ForceDelete()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
