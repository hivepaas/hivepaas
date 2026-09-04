package useruc

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

// projectAccessResourceTypes are the resource types a project grant is stored
// under. The project level is legacy - grants are written per env now - but old
// rows must still be cleared when they are replaced.
var projectAccessResourceTypes = []base.ResourceType{
	base.ResourceTypeProject,
	base.ResourceTypeProjectEnv,
}

// authorizeAccessChanges checks the grants built into persistingData against the
// acting user's own permissions, and fills in the rows they replace.
//
// The replaced set is deliberately not "everything the target has": a non-admin
// must not wipe grants they cannot see, so only rows they could revoke are
// cleared and the rest survive the update untouched.
func (uc *UC) authorizeAccessChanges(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	target *entity.User,
	resourceTypes []base.ResourceType,
	persistingData *userservice.PersistingUserData,
) error {
	if len(resourceTypes) == 0 {
		return nil
	}

	currentAccesses := make([]*entity.ACLPermission, 0, len(target.Accesses))
	for _, access := range target.Accesses {
		if slices.Contains(resourceTypes, access.ResourceType) {
			currentAccesses = append(currentAccesses, access)
		}
	}

	replaceable, err := uc.permissionManager.AuthorizeAccessChanges(ctx, db, auth,
		persistingData.UpsertingAccesses, currentAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, access := range replaceable {
		persistingData.DeletingAccesses = append(persistingData.DeletingAccesses,
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    target.ID,
				ResourceType: access.ResourceType,
				ResourceID:   access.ResourceID,
			})
	}
	return nil
}

// accessResourceTypesToReplace lists the resource types an operation replaces,
// based on which access lists the request carries.
func accessResourceTypesToReplace(replaceModules, replaceProjects bool) []base.ResourceType {
	var resourceTypes []base.ResourceType
	if replaceModules {
		resourceTypes = append(resourceTypes, base.ResourceTypeModule)
	}
	if replaceProjects {
		resourceTypes = append(resourceTypes, projectAccessResourceTypes...)
	}
	return resourceTypes
}
