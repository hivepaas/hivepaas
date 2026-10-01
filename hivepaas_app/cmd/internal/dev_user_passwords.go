package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

// DevResetUserPasswords gives every user the password dev mode names, each time
// a development install starts. The dev server is reseeded on every deploy, and
// the seed's password is in a public repository; this replaces it with one the
// deploy is given.
//
// A password it cannot set stops the start: the app is then not serving the
// seed's password, which is the point. Nothing happens outside dev mode, or
// with no password named.
func DevResetUserPasswords(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	userRepo repository.UserRepo,
	userService userservice.Service,
	logger logging.Logger,
) {
	password := cfg.DevMode.UserPassword
	if !cfg.DevMode.Enabled || password == "" || cfg.RunMode == config.RunModeUpdater {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			count, err := resetUserPasswords(ctx, db, userRepo, userService, password)
			if err != nil {
				return fmt.Errorf("dev mode: failed to reset the users' passwords: %w", err)
			}
			logger.Infof("dev mode: the password of %d users was reset", count)
			return nil
		},
	})
}

func resetUserPasswords(
	ctx context.Context,
	db database.IDB,
	userRepo repository.UserRepo,
	userService userservice.Service,
	password string,
) (int, error) {
	users, _, err := userRepo.List(ctx, db, nil)
	if err != nil {
		return 0, fmt.Errorf("listing the users: %w", err)
	}
	for _, user := range users {
		// The strength rules apply as to anyone's: a weak shared password is
		// refused here rather than handed to every account.
		if err = userService.ChangePassword(user, password, userservice.SkipCheckingCurrentPassword); err != nil {
			return 0, fmt.Errorf("user %s: %w", user.Username, err)
		}
		user.UpdatedAt = timeutil.NowUTC()
		if err = userRepo.Update(ctx, db, user, bunex.UpdateColumns("password", "updated_at")); err != nil {
			return 0, fmt.Errorf("user %s: %w", user.Username, err)
		}
	}
	return len(users), nil
}
