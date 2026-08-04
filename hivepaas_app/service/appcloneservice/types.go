package appcloneservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type AppCloneReq struct {
	SrcProject    *entity.Project
	SrcApp        *entity.App
	TargetProject *entity.Project

	OnCloneApp     func(targetApp, srcApp *entity.App) error
	OnCloneSetting func(targetApp *entity.App, s *entity.Setting) (*entity.Setting, error)
	OnCloneService func(targetSvc, srcSvc *swarm.Service) error
}

type AppCloneResp struct {
	TargetApp     *entity.App
	TargetService *swarm.Service
	OnCleanup     func(error) error
}
