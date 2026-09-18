package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type ResLinkRepo interface {
	Get(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstID string, opts ...bunex.SelectQueryOption) (*entity.ResLink, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ResLink, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, resLink *entity.ResLink, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error

	DeleteAllBySourceIDs(ctx context.Context, db database.IDB, sourceType base.ResourceType, sourceIDs []string,
		opts ...bunex.DeleteQueryOption) error
	// DeleteAllByScope deletes every link an object owns, its settings' links
	// included. Deleting an object uses this rather than DeleteAllBySourceIDs,
	// which reaches only links the object is itself the source of.
	DeleteAllByScope(ctx context.Context, db database.IDB, scope base.ObjectScopeType, objectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type resLinkRepo struct {
}

func NewResLinkRepo() ResLinkRepo {
	return &resLinkRepo{}
}

func (repo *resLinkRepo) Get(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
	dstType base.ResourceType, dstID string, opts ...bunex.SelectQueryOption) (*entity.ResLink, error) {
	resLink := &entity.ResLink{}
	query := db.NewSelect().Model(resLink).
		Where("res_link.src_id = ?", srcID).
		Where("res_link.dst_id = ?", dstID)
	if srcType != "" {
		query = query.Where("res_link.src_type = ?", srcType)
	}
	if dstType != "" {
		query = query.Where("res_link.dst_type = ?", dstType)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if resLink == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("ResLink").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resLink, nil
}

func (repo *resLinkRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ResLink, *basedto.PagingMeta, error) {
	var resLinks []*entity.ResLink
	query := db.NewSelect().Model(&resLinks)
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
	return resLinks, pagingMeta, nil
}

func (repo *resLinkRepo) Insert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.ResLink{resLink}, opts...)
}

func (repo *resLinkRepo) InsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
	opts ...bunex.InsertQueryOption) error {
	if len(resLinks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&resLinks)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *resLinkRepo) Upsert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.ResLink{resLink}, conflictCols, updateCols, opts...)
}

func (repo *resLinkRepo) UpsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(resLinks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&resLinks)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *resLinkRepo) DeleteAllBySourceIDs(ctx context.Context, db database.IDB,
	sourceType base.ResourceType, sourceIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.ResLink)(nil)).
		Where("res_link.src_type = ?", sourceType).
		Where("res_link.src_id IN (?)", bun.List(sourceIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// DeleteAllByScope deletes the links an object owns: the ones it is the source
// of, and the ones any of its settings is.
//
// The settings are what makes this necessary. Every link HivePaaS writes today
// is written by a setting - the routing setting records the domains and ports an
// app answers at, and each setting records what it refers to - so an app's links
// cannot be reached through the app's own id at all. Deleting the app without
// them leaves its domains and ports recorded as taken and the certificates it
// used impossible to remove.
//
// The settings are read including deleted ones, so it does not matter whether
// this runs before or after they are deleted.
func (repo *resLinkRepo) DeleteAllByScope(ctx context.Context, db database.IDB,
	scope base.ObjectScopeType, objectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(objectIDs) == 0 {
		return nil
	}
	_, err := repo.deleteAllByScopeQuery(db, scope, objectIDs, opts...).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// deleteAllByScopeQuery is the query DeleteAllByScope runs, apart so that what it
// matches can be read without a database.
func (repo *resLinkRepo) deleteAllByScopeQuery(db database.IDB, scope base.ObjectScopeType,
	objectIDs []string, opts ...bunex.DeleteQueryOption) *bun.DeleteQuery {
	settingIDs := db.NewSelect().Model((*entity.Setting)(nil)).
		Column("id").
		WhereAllWithDeleted().
		Where("setting.scope = ?", scope).
		Where("setting.object_id IN (?)", bun.List(objectIDs))

	// One group, so that the two sources stay an alternative to each other rather
	// than to whatever the caller's options add.
	query := db.NewDelete().Model((*entity.ResLink)(nil)).
		WhereGroup(" AND ", func(q *bun.DeleteQuery) *bun.DeleteQuery {
			q = q.Where("res_link.src_type = ? AND res_link.src_id IN (?)",
				base.ResourceTypeSetting, settingIDs)
			if srcType := scopeSourceType(scope); srcType != "" {
				q = q.WhereOr("res_link.src_type = ? AND res_link.src_id IN (?)",
					srcType, bun.List(objectIDs))
			}
			return q
		})
	return bunex.ApplyDelete(query, opts...)
}

// scopeSourceType is the resource an object of a scope is, for the links it is
// itself the source of. Nothing writes those today, and deleting them is what
// the callers of this asked for before it existed.
func scopeSourceType(scope base.ObjectScopeType) base.ResourceType {
	switch scope {
	case base.ObjectScopeApp:
		return base.ResourceTypeApp
	case base.ObjectScopeProject:
		return base.ResourceTypeProject
	case base.ObjectScopeProjectEnv:
		return base.ResourceTypeProjectEnv
	case base.ObjectScopeUser:
		return base.ResourceTypeUser
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return ""
	default:
		return ""
	}
}

func (repo *resLinkRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.ResLink)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
