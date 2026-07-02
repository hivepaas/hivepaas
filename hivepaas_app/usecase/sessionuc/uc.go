package sessionuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	cacheLoginAttemptRepo cacherepository.LoginAttemptRepo
	cacheMfaPasscodeRepo  cacherepository.MFAPasscodeRepo
	db                    *database.DB

	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo
	settingRepo            repository.SettingRepo
	systemStatusRepo       repository.SystemStatusRepo
	userRepo               repository.UserRepo
	userTokenRepo          cacherepository.UserTokenRepo

	emailService emailservice.Service
	userService  userservice.Service

	permissionManager permission.Manager
}

func New(
	cacheLoginAttemptRepo cacherepository.LoginAttemptRepo,
	cacheMfaPasscodeRepo cacherepository.MFAPasscodeRepo,
	db *database.DB,

	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo,
	settingRepo repository.SettingRepo,
	systemStatusRepo repository.SystemStatusRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	emailService emailservice.Service,
	userService userservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		cacheLoginAttemptRepo: cacheLoginAttemptRepo,
		cacheMfaPasscodeRepo:  cacheMfaPasscodeRepo,
		db:                    db,

		loginTrustedDeviceRepo: loginTrustedDeviceRepo,
		settingRepo:            settingRepo,
		systemStatusRepo:       systemStatusRepo,
		userRepo:               userRepo,
		userTokenRepo:          userTokenRepo,

		emailService: emailService,
		userService:  userService,

		permissionManager: permissionManager,
	}
}
