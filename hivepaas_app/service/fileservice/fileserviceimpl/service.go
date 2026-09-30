package fileserviceimpl

import (
	agentfile "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	fileRepo repository.FileRepo,
	settingRepo repository.SettingRepo,
	dockerManager docker.Manager,
	agentService agentservice.Service,
) fileservice.Service {
	return &service{
		fileRepo:      fileRepo,
		settingRepo:   settingRepo,
		dockerManager: dockerManager,
		agentService:  agentService,
		agentFiles:    agentfile.NewFileServiceClient,
	}
}

type service struct {
	fileRepo      repository.FileRepo
	settingRepo   repository.SettingRepo
	dockerManager docker.Manager
	agentService  agentservice.Service
	// agentFiles connects to the file service of the agent at an address.
	agentFiles func(agentAddr string) (agentfile.FileServiceClient, error)
}
