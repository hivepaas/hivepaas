package userserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadNotificationUsers(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
	loadMembers bool,
	loadOwners bool,
	loadAdmins bool,
) (map[string]*entity.User, error) {
	if !loadMembers && !loadOwners && !loadAdmins {
		return nil, nil
	}
	userIDs := make([]string, 0, 10) //nolint:mnd

	if loadMembers && project != nil {
		perms, err := s.permissionManager.LoadProjectAccessUsers(ctx, db, project.ID, nil)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for _, perm := range perms {
			userIDs = append(userIDs, perm.SubjectID)
		}
	}

	if loadOwners && project != nil && project.OwnerID != "" {
		userIDs = append(userIDs, project.OwnerID)
	}

	userMap, err := s.LoadUsersEx(ctx, db, false,
		bunex.SelectWhere("\"user\".id IN (?)", bunex.List(userIDs)),
		bunex.SelectWhereOrIf(loadAdmins, "\"user\".role = ?", base.UserRoleAdmin),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return userMap, nil
}
