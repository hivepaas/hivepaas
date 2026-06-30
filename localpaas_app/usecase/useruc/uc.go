package useruc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
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
