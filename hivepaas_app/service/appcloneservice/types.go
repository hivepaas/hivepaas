package appcloneservice

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type AppCloneReq struct {
	SrcApp *entity.App

	// CloneSettings settings to clone an app, can be nil if passing the custom callbacks
	CloneSettings *entity.AppCloneSettings
	// Custom callbacks to override settings
	OnCloneApp     func(destApp, srcApp *entity.App) error
	OnCloneSetting func(destApp, srcApp *entity.App, s *entity.Setting) (*entity.Setting, error)
	OnCloneService func(destApp, srcApp *entity.App, destSvc, srcSvc *swarm.Service) error
	OnCloneVolumes func(destApp, srcApp *entity.App, srcMount []mount.Mount) ([]mount.Mount, error)

	LogStore *tasklog.Store
}

type AppCloneResp struct {
	TargetApp     *entity.App
	TargetService *swarm.Service
	OnCleanup     func(error) error
}
