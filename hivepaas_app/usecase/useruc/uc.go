package useruc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
	settingRepo   repository.SettingRepo
	userRepo      repository.UserRepo
	userTokenRepo cacherepository.UserTokenRepo

	auditService auditservice.Service
	emailService emailservice.Service
	userService  userservice.Service

	permissionManager permission.Manager
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
	settingRepo repository.SettingRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	auditService auditservice.Service,
	emailService emailservice.Service,
	userService userservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
		settingRepo:   settingRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,

		auditService: auditService,
		emailService: emailService,
		userService:  userService,

		permissionManager: permissionManager,
	}
}
