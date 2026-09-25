package appcloneservice

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type AppCloneReq struct {
	// TaskExecData is the clone's own: the task it is logged against, and the
	// callbacks that reach the end of the transaction it runs in. A caller that
	// is a task passes its own; a caller doing this inside another task - a
	// preview - passes queue.TaskExecData.SubTask, which names the clone without
	// losing where the transaction ends.
	*queue.TaskExecData

	SrcApp *entity.App

	// CloneSettings settings to clone an app, can be nil if passing the custom callbacks
	CloneSettings *entity.AppCloneSettings

	// DropGatedMounts leaves out setting mounts with a gated part - a private
	// key, a password: the clone's requester may not reveal them.
	DropGatedMounts bool

	// Custom callbacks to override settings
	OnCloneStart   func(req *AppCloneReq) error
	OnCloneApp     func(destApp, srcApp *entity.App) error
	OnCloneSetting func(destApp, srcApp *entity.App, s *entity.Setting) (*entity.Setting, error)
	OnCloneService func(destApp, srcApp *entity.App, destSvc, srcSvc *swarm.Service) error
	OnCloneVolumes func(destApp, srcApp *entity.App, srcMount []mount.Mount) ([]mount.Mount, error)
}

type AppCloneResp struct {
	TargetApp     *entity.App
	TargetService *swarm.Service
	OnCleanup     func(error) error
}
