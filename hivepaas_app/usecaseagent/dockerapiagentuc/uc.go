package dockerapiagentuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	syncInterval    = 30 * time.Second
	collectInterval = 10 * time.Minute
)

// UC serves, on this node, the Docker API of every app that has access, and
// removes what the apps' children leave behind.
type UC struct {
	logger           logging.Logger
	db               *database.DB
	dockerManager    docker.Manager
	dockerAPIService dockerapiservice.Service
	host             *socketHost
}

func New(
	logger logging.Logger,
	db *database.DB,
	dockerManager docker.Manager,
	dockerAPIService dockerapiservice.Service,
) (*UC, error) {
	upstream, err := newUpstream()
	if err != nil {
		return nil, err
	}
	return &UC{
		logger:           logger,
		db:               db,
		dockerManager:    dockerManager,
		dockerAPIService: dockerAPIService,
		host:             newSocketHost(logger, dockerManager, upstream),
	}, nil
}

// Sync serves exactly the apps that have access now, and says how many this
// node serves. When the database cannot be read, nothing changes.
func (uc *UC) Sync(ctx context.Context) (int, error) {
	policies, err := uc.dockerAPIService.Policies(ctx, uc.db)
	if err != nil {
		return len(uc.host.served()), hperrors.Wrap(err)
	}
	err = uc.host.reconcile(ctx, policies)
	return len(uc.host.served()), err
}

// RemoveApp stops serving an app and removes what its children left on this
// node.
func (uc *UC) RemoveApp(ctx context.Context, appID string) (Removed, error) {
	uc.host.closeApp(appID)
	s := &sweep{docker: uc.dockerManager, now: time.Now(), only: appID,
		gone: func(id string) bool { return id == appID }}
	return s.run(ctx)
}

// Collect removes what apps' children left behind on this node.
func (uc *UC) Collect(ctx context.Context) (Removed, error) {
	policies, err := uc.dockerAPIService.Policies(ctx, uc.db)
	if err != nil {
		// Without knowing who has access, nothing is anybody's to remove.
		return Removed{}, hperrors.Wrap(err)
	}
	access := make(map[string]bool, len(policies))
	for _, policy := range policies {
		access[policy.AppID] = true
	}
	s := &sweep{docker: uc.dockerManager, now: time.Now(), ageOut: true,
		gone: func(appID string) bool { return !access[appID] }}
	return s.run(ctx)
}

// Run syncs at once and then on a timer, and collects on a longer one, until
// ctx ends.
func (uc *UC) Run(ctx context.Context) {
	uc.syncAndLog(ctx)
	syncTicker := time.NewTicker(syncInterval)
	defer syncTicker.Stop()
	collectTicker := time.NewTicker(collectInterval)
	defer collectTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-syncTicker.C:
			uc.syncAndLog(ctx)
		case <-collectTicker.C:
			uc.collectAndLog(ctx)
		}
	}
}

// Close stops serving every app.
func (uc *UC) Close() {
	uc.host.closeAll()
}

func (uc *UC) syncAndLog(ctx context.Context) {
	if _, err := uc.Sync(ctx); err != nil {
		uc.logger.Errorf("docker api: sync: %v", err)
	}
}

func (uc *UC) collectAndLog(ctx context.Context) {
	removed, err := uc.Collect(ctx)
	if err != nil {
		uc.logger.Errorf("docker api: collect: %v", err)
	}
	if removed != (Removed{}) {
		uc.logger.Infof("docker api: removed %d containers, %d networks and %d volumes left behind",
			removed.Containers, removed.Networks, removed.Volumes)
	}
}
