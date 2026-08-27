package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (p *manager) DeleteProjectAccesses(
	ctx context.Context,
	db database.IDB,
	projectID string,
	extraOpts ...bunex.DeleteQueryOption,
) error {
	opts := []bunex.DeleteQueryOption{
		bunex.DeleteWhere("acl_permission.subj_type = ?", string(base.SubjectTypeUser)),
		bunex.DeleteWhereGroup(
			// all ACLs of project
			bunex.DeleteWhere("(acl_permission.res_type = ? AND acl_permission.res_id = ?)",
				string(base.ResourceTypeProject), projectID),
			// all ACLs of all project's envs
			bunex.DeleteWhereOr("(acl_permission.res_type = ? AND acl_permission.res_id LIKE ?)",
				string(base.ResourceTypeProjectEnv), projectID+":%"),
		),
	}
	opts = append(opts, extraOpts...)

	err := p.aclPermissionRepo.Delete(ctx, db, opts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
