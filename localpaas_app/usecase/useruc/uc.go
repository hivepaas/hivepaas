package useruc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

type UC struct {
	db           *database.DB
	userRepo     repository.UserRepo
	userService  userservice.Service
	emailService emailservice.Service
}

func New(
	db *database.DB,
	userRepo repository.UserRepo,
	userService userservice.Service,
	emailService emailservice.Service,
) *UC {
	return &UC{
		db:           db,
		userRepo:     userRepo,
		userService:  userService,
		emailService: emailService,
	}
}
