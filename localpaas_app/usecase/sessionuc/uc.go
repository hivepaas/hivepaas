package sessionuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

type UC struct {
	db                     *database.DB
	userRepo               repository.UserRepo
	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo
	settingRepo            repository.SettingRepo
	userTokenRepo          cacherepository.UserTokenRepo
	cacheMfaPasscodeRepo   cacherepository.MFAPasscodeRepo
	cacheLoginAttemptRepo  cacherepository.LoginAttemptRepo
	userService            userservice.Service
	emailService           emailservice.Service
	permissionManager      permission.Manager
}

func New(
	db *database.DB,
	userRepo repository.UserRepo,
	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo,
	settingRepo repository.SettingRepo,
	userTokenRepo cacherepository.UserTokenRepo,
	cacheMfaPasscodeRepo cacherepository.MFAPasscodeRepo,
	cacheLoginAttemptRepo cacherepository.LoginAttemptRepo,
	userService userservice.Service,
	emailService emailservice.Service,
	permissionManager permission.Manager,
) *UC {
	return &UC{
		db:                     db,
		userRepo:               userRepo,
		loginTrustedDeviceRepo: loginTrustedDeviceRepo,
		settingRepo:            settingRepo,
		userTokenRepo:          userTokenRepo,
		cacheMfaPasscodeRepo:   cacheMfaPasscodeRepo,
		cacheLoginAttemptRepo:  cacheLoginAttemptRepo,
		userService:            userService,
		emailService:           emailService,
		permissionManager:      permissionManager,
	}
}
