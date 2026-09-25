package settingmountserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	resLinkRepo repository.ResLinkRepo
	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	dockerManager docker.Manager
	taskQueue     queue.TaskQueue
	logger        logging.Logger

	// removalRetryDelay is how long Sweep waits before asking again to remove
	// an object a service still holds.
	removalRetryDelay time.Duration

	// The seams below are what tests replace: bunex options are opaque closures,
	// which a test double of the repositories cannot read.
	rotationKey func() []byte
	loadEntries func(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error)
	loadSources func(ctx context.Context, db database.IDB, app *entity.App, ids []string) ([]*entity.Setting, error)
	loadReaders func(ctx context.Context, db database.IDB, sourceIDs []string) ([]string, error)
}

// loadEntriesFromRepo is the app's own entries, whatever their status: an entry
// is never inherited.
func (s *service) loadEntriesFromRepo(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error) {
	entries, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
		bunex.SelectWhere("setting.object_id = ?", appID),
	)
	return entries, hperrors.Wrap(err)
}

// loadSourcesFromRepo is the active sources among ids that the app's scope sees.
func (s *service) loadSourcesFromRepo(
	ctx context.Context, db database.IDB, app *entity.App, ids []string,
) ([]*entity.Setting, error) {
	sources, err := s.settingRepo.ListByIDs(ctx, db, app.GetObjectScope(), ids, true)
	return sources, hperrors.Wrap(err)
}
