package repository

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type ACLPermissionRepo interface {
	ListByResources(ctx context.Context, db database.IDB, resources []*base.PermissionResource,
		opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error)
	ListByUsers(ctx context.Context, db database.IDB, userIDs []string,
		opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, permissions []*entity.ACLPermission,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error

	DeleteByResources(ctx context.Context, db database.IDB, resources []*base.PermissionResource,
		opts ...bunex.DeleteQueryOption) error
	DeleteByUsers(ctx context.Context, db database.IDB, userIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type aclPermissionRepo struct {
}

func NewACLPermissionRepo() ACLPermissionRepo {
	return &aclPermissionRepo{}
}

func (repo *aclPermissionRepo) ListByResources(ctx context.Context, db database.IDB,
	resources []*base.PermissionResource, opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error) {
	if len(resources) == 0 {
		return nil, nil
	}
	var condition string
	args := make([]any, 0, len(resources)*2) //nolint:mnd
	for _, res := range resources {
		subjectCol := "subject_id"
		subjectArg := res.SubjectID
		if res.SubjectID == "" {
			subjectCol = "subject_type"
			subjectArg = string(res.SubjectType)
		}
		resCol := "resource_id"
		resArg := res.ResourceID
		if res.ResourceID == "" {
			resCol = "resource_type"
			resArg = string(res.ResourceType)
		}
		if condition == "" {
			condition = fmt.Sprintf("(%s,%s) = (?,?)", subjectCol, resCol)
		} else {
			condition = fmt.Sprintf("%s OR (%s,%s) = (?,?)", condition, subjectCol, resCol)
		}
		args = append(args, subjectArg, resArg)
	}

	var permissions []*entity.ACLPermission
	query := db.NewSelect().Model(&permissions).Where(condition, args...)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return permissions, nil
}

func (repo *aclPermissionRepo) ListByUsers(ctx context.Context, db database.IDB, userIDs []string,
	opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var permissions []*entity.ACLPermission
	query := db.NewSelect().Model(&permissions).Where("subject_id IN (?)", bun.List(userIDs))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return permissions, nil
}

func (repo *aclPermissionRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ACLPermission, *basedto.PagingMeta, error) {
	var acls []*entity.ACLPermission
	query := db.NewSelect().Model(&acls)
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

	return acls, pagingMeta, nil
}

func (repo *aclPermissionRepo) UpsertMulti(ctx context.Context, db database.IDB, permissions []*entity.ACLPermission,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(permissions) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&permissions)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *aclPermissionRepo) DeleteByResources(ctx context.Context, db database.IDB,
	resources []*base.PermissionResource, opts ...bunex.DeleteQueryOption) error {
	if len(resources) == 0 {
		return nil
	}
	var condition string
	args := make([]any, 0, len(resources)*2) //nolint:mnd
	for _, res := range resources {
		subjectCol := "subject_id"
		subjectArg := res.SubjectID
		if res.SubjectID == "" {
			subjectCol = "subject_type"
			subjectArg = string(res.SubjectType)
		}
		resCol := "resource_id"
		resArg := res.ResourceID
		if res.ResourceID == "" {
			resCol = "resource_type"
			resArg = string(res.ResourceType)
		}
		if condition == "" {
			condition = fmt.Sprintf("(%s,%s) = (?,?)", subjectCol, resCol)
		} else {
			condition = fmt.Sprintf("%s OR (%s,%s) = (?,?)", condition, subjectCol, resCol)
		}
		args = append(args, subjectArg, resArg)
	}

	query := db.NewDelete().Model((*entity.ACLPermission)(nil))
	if condition != "" {
		query = query.Where(condition, args...)
	}
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *aclPermissionRepo) DeleteByUsers(ctx context.Context, db database.IDB, userIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.ACLPermission)(nil)).
		Where("subject_id IN (?)", bun.List(userIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *aclPermissionRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.ACLPermission)(nil)).ForceDelete()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
