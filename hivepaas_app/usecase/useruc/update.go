package useruc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) UpdateUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdateUserReq,
) (*userdto.UpdateUserResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		userData := &userUpdateData{}
		err := uc.loadUserDataForUpdate(ctx, db, auth, req, userData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Read before prepareUpdatingUserData writes over them.
		before := userData.snapshot()

		persistingData := &userservice.PersistingUserData{}
		uc.prepareUpdatingUserData(req, userData, persistingData)

		err = uc.authorizeAccessChanges(ctx, db, auth, userData.User,
			accessResourceTypesToReplace(req.ModuleAccesses != nil, req.Capabilities != nil,
				req.ProjectAccesses != nil),
			persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.cascadeRevokedCapabilities(ctx, db, auth, userData.User, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Revoke target user's JWT, user needs to re-login
		err = uc.userTokenRepo.DelAll(ctx, req.ID)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if err = uc.userService.PersistUserData(ctx, db, persistingData); err != nil {
			return hperrors.Wrap(err)
		}

		// The privilege-bearing fields go in with their values - a role, a status
		// and a security option are not secrets, and "member to admin" is the
		// whole of what a reader wants from the entry. Which grant lists were
		// replaced is said by name only: their contents are the permission model,
		// not something an audit row has to carry.
		user := userData.User
		return uc.recordUserChange(ctx, db, auth, base.AuditLogTypeUserUpdate,
			auditSectionAccount, user, auditdetail.New().
				Compare("role", before.Role, user.Role).
				Compare("status", before.Status, user.Status).
				Compare("securityOption", before.SecurityOption, user.SecurityOption).
				Set("moduleAccessesReplaced", req.ModuleAccesses != nil).
				Set("capabilitiesReplaced", req.Capabilities != nil).
				Set("projectAccessesReplaced", req.ProjectAccesses != nil))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.UpdateUserResp{}, nil
}

type userUpdateData struct {
	User *entity.User
}

// userSnapshot is what the record compares against: the fields that decide what
// an account may do, as they were before the request touched them.
type userSnapshot struct {
	Role           base.UserRole
	Status         base.UserStatus
	SecurityOption base.UserSecurityOption
}

func (data *userUpdateData) snapshot() userSnapshot {
	return userSnapshot{
		Role:           data.User.Role,
		Status:         data.User.Status,
		SecurityOption: data.User.SecurityOption,
	}
}

func (uc *UC) loadUserDataForUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.UpdateUserReq,
	data *userUpdateData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, req.ID,
		// The current grants decide what this update may replace.
		bunex.SelectRelation("Accesses"),
		bunex.SelectFor("UPDATE OF \"user\""),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.User = user

	// If username changes, need to verify the uniqueness
	if req.Username != "" && req.Username != user.Username {
		conflictUser, err := uc.userRepo.GetByUsername(ctx, db, req.Username)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if conflictUser != nil {
			return hperrors.Wrap(hperrors.ErrUsernameUnavailable).
				WithMsgLog("user '%s' already exists", req.Username)
		}
	}

	// If email changes, need to verify the uniqueness
	if req.Email != "" && req.Email != user.Email {
		conflictUser, err := uc.userRepo.GetByEmail(ctx, db, req.Email)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if conflictUser != nil {
			return hperrors.Wrap(hperrors.ErrEmailUnavailable).
				WithMsgLog("email '%s' already exists", req.Email)
		}
	}

	if req.Role != nil {
		if base.RoleCmp(auth.User.Role, *req.Role) < 0 {
			return hperrors.Wrap(hperrors.ErrForbidden).
				WithMsgLog("you are not allowed to set a role higher than yours")
		}
	}

	if err = ensureAdminSecurityOption(req, user); err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

// ensureAdminSecurityOption refuses an update that would leave an admin account
// authenticating with a password and nothing else.
//
// Read against the result of the update rather than the request, because either
// half can be left out and the dangerous combinations arrive both ways round:
// promoting a password-only member sends a role and no security option, and
// weakening an admin sends a security option and no role.
//
// Only when the request moves one of them. An account whose combination predates
// this rule - the bootstrap admin is seeded password-only, see
// userserviceimpl.initAdminUser - would otherwise have every edit through the
// user form refused, including the one that disables it in a hurry. Refusing to
// disable a compromised admin because their authentication is weak is the wrong
// way round, and the header's disable button sends a status and nothing else, so
// it stays clear of this either way.
func ensureAdminSecurityOption(req *userdto.UpdateUserReq, user *entity.User) error {
	role, option := user.Role, user.SecurityOption
	var changed bool
	if req.Role != nil && *req.Role != role {
		role, changed = *req.Role, true
	}
	if req.SecurityOption != nil && *req.SecurityOption != option {
		option, changed = *req.SecurityOption, true
	}

	if !changed || base.SecurityOptionAllowedForRole(role, option) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrUserAdminSecurityOptionWeak).
		WithMsgLog("an admin account cannot use the '%s' security option", option)
}

func (uc *UC) prepareUpdatingUserData(
	req *userdto.UpdateUserReq,
	updateData *userUpdateData,
	persistingData *userservice.PersistingUserData,
) {
	timeNow := timeutil.NowUTC()
	user := updateData.User

	user.UpdatedAt = timeNow
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Position != nil {
		user.Position = *req.Position
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Notes != nil {
		user.Notes = *req.Notes
	}
	oldSecurityOption := user.SecurityOption
	if req.SecurityOption != nil {
		user.SecurityOption = *req.SecurityOption
	}
	if req.AccessExpireAt != nil {
		user.AccessExpireAt = *req.AccessExpireAt
	}

	switch user.Status {
	case base.UserStatusActive:
		// User needs to set up 2FA authentication, set user status to `pending`
		if user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
			user.Status = base.UserStatusPending
		}
	case base.UserStatusPending:
		// Look like admin changes user setting from `2FA` back to `password-only`
		if oldSecurityOption == base.UserSecurityPassword2FA &&
			user.SecurityOption == base.UserSecurityPasswordOnly && user.Password != "" {
			user.Status = base.UserStatusActive
		}
	case base.UserStatusDisabled:
		// Do nothing
	}

	persistingData.UpsertingUsers = append(persistingData.UpsertingUsers, user)

	if req.ModuleAccesses != nil {
		uc.preparePersistingUserModuleAccesses(user, req.ModuleAccesses, timeNow, persistingData)
	}
	if req.Capabilities != nil {
		uc.preparePersistingUserCapabilities(user, req.Capabilities, timeNow, persistingData)
	}
	if req.ProjectAccesses != nil {
		uc.preparePersistingUserProjectAccesses(user, req.ProjectAccesses, timeNow, persistingData)
	}
}
