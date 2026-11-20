package clusterserviceimpl

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	dockerManager docker.Manager,
) clusterservice.Service {
	return &service{
		appRepo:     appRepo,
		settingRepo: settingRepo,

		dockerManager: dockerManager,
	}
}

type service struct {
	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	dockerManager docker.Manager

	// taskPollInterval is how often ServiceRemoveAndWait reads the tasks; zero
	// is a second. Tests shorten it.
	taskPollInterval time.Duration
}
