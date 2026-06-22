package appcopyservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type AppCopyReq struct {
	SrcProject    *entity.Project
	SrcApp        *entity.App
	TargetProject *entity.Project

	CopyApp     func(targetApp, srcApp *entity.App) error
	CopySetting func(targetApp *entity.App, s *entity.Setting) (*entity.Setting, error)
	CopyService func(targetSvc, srcSvc *swarm.Service) error
}

type AppCopyResp struct {
	TargetApp     *entity.App
	TargetService *swarm.Service
	CleanupFunc   func(error) error
}
