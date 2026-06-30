package userserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

func New(
	binObjectRepo repository.BinObjectRepo,
	fileRepo repository.FileRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	permissionManager permission.Manager,
) userservice.Service {
	return &service{
		binObjectRepo: binObjectRepo,
		fileRepo:      fileRepo,
		resLinkRepo:   resLinkRepo,
		settingRepo:   settingRepo,
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,

		permissionManager: permissionManager,
	}
}

type service struct {
	binObjectRepo repository.BinObjectRepo
	fileRepo      repository.FileRepo
	resLinkRepo   repository.ResLinkRepo
	settingRepo   repository.SettingRepo
	taskRepo      repository.TaskRepo
	userRepo      repository.UserRepo
	userTokenRepo cacherepository.UserTokenRepo

	permissionManager permission.Manager
}
