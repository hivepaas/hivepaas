package useruc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
	userRepo      repository.UserRepo
	userTokenRepo cacherepository.UserTokenRepo

	emailService emailservice.Service
	userService  userservice.Service
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	emailService emailservice.Service,
	userService userservice.Service,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,

		emailService: emailService,
		userService:  userService,
	}
}
