package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db            *database.DB
	dockerManager docker.Manager

	settingRepo repository.SettingRepo
	tagRepo     repository.TagRepo

	agentService    agentservice.Service
	nodeExecService nodeexecservice.Service
	settingService  settingservice.Service

	// agentAddr and openRepoServer reach a node's agent and run a repository
	// server there; tests replace them.
	agentAddr      func(ctx context.Context, nodeID, nodeLabel string) (string, error)
	openRepoServer func(ctx context.Context, agentAddr string, req *reposerveragentuc.RunReq) (*openedRepoServer, error)
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	settingRepo repository.SettingRepo,
	tagRepo repository.TagRepo,

	agentService agentservice.Service,
	nodeExecService nodeexecservice.Service,
	settingService settingservice.Service,
) backupreposervice.Service {
	svc := &service{
		db:            db,
		dockerManager: dockerManager,

		settingRepo: settingRepo,
		tagRepo:     tagRepo,

		agentService:    agentService,
		nodeExecService: nodeExecService,
		settingService:  settingService,

		openRepoServer: openRepoServerOnAgent,
	}
	svc.agentAddr = svc.agentAddrForNode
	return svc
}
