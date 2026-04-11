package userservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/appentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type Service interface {
	InitAdminUser(ctx context.Context, db database.IDB) (err error)

	LoadUser(ctx context.Context, db database.IDB, userID string) (*entity.User, error)
	LoadUsers(ctx context.Context, db database.IDB, userIDs []string, errorIfUnavail bool) (
		userMap map[string]*entity.User, err error)
	LoadUsersEx(ctx context.Context, db database.IDB, errorIfUnavail bool,
		loadOpts ...bunex.SelectQueryOption) (map[string]*entity.User, error)
	LoadUserByEmail(ctx context.Context, db database.IDB, email string) (*entity.User, error)
	LoadUsersByEmails(ctx context.Context, db database.IDB, emails []string, errorIfUnavail bool) (
		userMap map[string]*entity.User, err error)

	PersistUserData(ctx context.Context, db database.IDB, persistingData *PersistingUserData) error
	SaveUserPhoto(_ context.Context, user *entity.User, data []byte, fileExt string) error

	ChangePassword(user *entity.User, newPassword, currPassword string) error
	VerifyPassword(user *entity.User, password string) error
	CheckPasswordStrength(password string) error

	GenerateMFAToken(userID string, mfaType base.MFAType, trustedDeviceID string) (string, error)
	ParseMFAToken(token string) (*appentity.MFATokenClaims, error)
	GenerateMFATotpSetupToken(userID string, toptSecret string) (string, error)
	ParseMFATotpSetupToken(token string) (*appentity.MFATotpSetupTokenClaims, error)
	GenerateUserInviteToken(userID string) (string, error)
	ParseUserInviteToken(token string) (*appentity.UserInviteTokenClaims, error)
	GeneratePasswordResetToken(userID string) (string, error)
	ParsePasswordResetToken(token string) (*appentity.PasswordResetTokenClaims, error)

	// Project users
	LoadProjectUsers(ctx context.Context, db database.IDB, project *entity.Project,
		loadMembers bool, loadOwners bool, loadAdmins bool) (map[string]*entity.User, error)
}
