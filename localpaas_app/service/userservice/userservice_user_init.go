package userservice

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	envAdminEmail    = "LP_ADMIN_EMAIL"
	envAdminUsername = "LP_ADMIN_USERNAME"
	envAdminPassword = "LP_ADMIN_PASSWORD"
)

func (s *userService) InitAdminUser(
	ctx context.Context,
	db database.IDB,
) (cleanupFunc func(), err error) {
	email := os.Getenv(envAdminEmail)
	password := os.Getenv(envAdminPassword)
	if email == "" || password == "" {
		return nil, apperrors.NewMissing("Email or password is missing")
	}
	username := gofn.Coalesce(os.Getenv(envAdminUsername), strings.Split(email, "@")[0])

	timeNow := timeutil.NowUTC()
	user := &entity.User{
		ID:             gofn.Must(ulid.NewStringULID()),
		Email:          email,
		Username:       username,
		Role:           base.UserRoleAdmin,
		Status:         base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
	}

	user.Password, err = s.createPasswordHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}

	err = s.userRepo.Upsert(ctx, db, user, entity.UserUpsertingConflictCols, entity.UserUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return func() {
		// Clear sensitive ENV vars
		_ = os.Unsetenv(envAdminEmail)
		_ = os.Unsetenv(envAdminUsername)
		_ = os.Unsetenv(envAdminPassword)
	}, nil
}
