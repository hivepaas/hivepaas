package userserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

func New(
	userRepo repository.UserRepo,
	settingRepo repository.SettingRepo,
	resLinkRepo repository.ResLinkRepo,
	fileRepo repository.FileRepo,
	taskRepo repository.TaskRepo,
	binObjectRepo repository.BinObjectRepo,
	userTokenRepo cacherepository.UserTokenRepo,
	permissionManager permission.Manager,
) userservice.Service {
	return &service{
		userRepo:          userRepo,
		settingRepo:       settingRepo,
		resLinkRepo:       resLinkRepo,
		fileRepo:          fileRepo,
		taskRepo:          taskRepo,
		binObjectRepo:     binObjectRepo,
		userTokenRepo:     userTokenRepo,
		permissionManager: permissionManager,
	}
}

type service struct {
	userRepo          repository.UserRepo
	settingRepo       repository.SettingRepo
	resLinkRepo       repository.ResLinkRepo
	fileRepo          repository.FileRepo
	taskRepo          repository.TaskRepo
	binObjectRepo     repository.BinObjectRepo
	userTokenRepo     cacherepository.UserTokenRepo
	permissionManager permission.Manager
}
