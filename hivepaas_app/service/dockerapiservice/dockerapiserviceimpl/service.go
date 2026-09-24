package dockerapiserviceimpl

import (
	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	settingRepo    repository.SettingRepo
	appRepo        repository.AppRepo
	networkService networkservice.Service
	agentService   agentservice.Service
	dockerManager  docker.Manager
	// agentClient reaches one node's agent.
	agentClient func(addr string) (dockerapiclient.DockerAPIServiceClient, error)
}

func New(
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	networkService networkservice.Service,
	agentService agentservice.Service,
	dockerManager docker.Manager,
) dockerapiservice.Service {
	return &service{
		settingRepo:    settingRepo,
		appRepo:        appRepo,
		networkService: networkService,
		agentService:   agentService,
		dockerManager:  dockerManager,
		agentClient:    dockerapiclient.NewDockerAPIServiceClient,
	}
}
