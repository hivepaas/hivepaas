package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (p *manager) UpdateACLPermissions(
	ctx context.Context,
	db database.IDB,
	perms []*entity.ACLPermission,
) error {
	err := p.aclPermissionRepo.UpsertMulti(ctx, db, perms,
		entity.ACLPermissionUpsertingConflictCols, entity.ACLPermissionUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (p *manager) RemoveACLPermissions(
	ctx context.Context,
	db database.IDB,
	perms []*base.PermissionResource,
) error {
	err := p.aclPermissionRepo.DeleteByResources(ctx, db, perms)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (p *manager) RemoveACLPermissionsBySubjects(
	ctx context.Context,
	db database.IDB,
	subjectType base.SubjectType,
	subjectIDs []string,
) error {
	err := p.aclPermissionRepo.DeleteBySubjects(ctx, db, subjectType, subjectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (p *manager) RemoveACLPermissionsOfUsers(
	ctx context.Context,
	db database.IDB,
	userIDs []string,
) error {
	err := p.aclPermissionRepo.DeleteBySubjects(ctx, db, base.SubjectTypeUser, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
