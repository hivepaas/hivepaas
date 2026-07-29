package userserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) DeleteUser(ctx context.Context, db database.IDB, user *entity.User) error {
	// Revoke target user's JWT, user can't access with the old token
	err := s.userTokenRepo.DelAll(ctx, user.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete ref resources in DB
	userIDs := []string{user.ID}

	// ACL permissions related to the user
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// User files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeUser, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeUser, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeUser, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByUsers(ctx, db, userIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// User photo
	if user.PhotoID != "" {
		err = s.binObjectRepo.DeleteByIDs(ctx, db, []string{user.PhotoID})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	user.DeletedAt = time.Now()
	err = s.userRepo.Update(ctx, db, user, bunex.UpdateColumns("deleted_at"))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
