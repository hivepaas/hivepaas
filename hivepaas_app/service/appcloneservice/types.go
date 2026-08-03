package appcloneservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type AppCloneReq struct {
	SrcProject    *entity.Project
	SrcApp        *entity.App
	TargetProject *entity.Project

	OnCopyApp     func(targetApp, srcApp *entity.App) error
	OnCopySetting func(targetApp *entity.App, s *entity.Setting) (*entity.Setting, error)
	OnCopyService func(targetSvc, srcSvc *swarm.Service) error
}

type AppCloneResp struct {
	TargetApp     *entity.App
	TargetService *swarm.Service
	OnCleanup     func(error) error
}
