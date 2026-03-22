package userservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (s *userService) LoadProjectUsers(
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
		accesses, err := s.permissionManager.LoadObjectAccesses(ctx, db, &permission.AccessCheck{
			SubjectType:  base.SubjectTypeUser,
			ResourceType: base.ResourceTypeProject,
			ResourceID:   project.ID,
			Action:       base.ActionTypeRead,
		}, false)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for _, access := range accesses {
			userIDs = append(userIDs, access.SubjectID)
		}
	}

	if loadOwners && project != nil && project.OwnerID != "" {
		userIDs = append(userIDs, project.OwnerID)
	}

	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("\"user\".id IN (?)", bunex.In(userIDs)),
	}
	if loadAdmins {
		opts = append(opts, bunex.SelectWhereOr("\"user\".role = ?", base.UserRoleAdmin))
	}

	userMap, err := s.LoadUsersEx(ctx, db, false, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return userMap, nil
}
