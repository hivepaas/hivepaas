package attentionserviceimpl

import (
	"context"
	"sync"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// cacheTTL is how long one reading of the cluster serves. Every user's home
// page asks, and the answer is the same for all of them before it is narrowed:
// reading every service, task and node for each would cost the cluster more
// than the page is worth.
const cacheTTL = 30 * time.Second

type service struct {
	appRepo       repository.AppRepo
	dockerManager docker.Manager

	mu      sync.Mutex
	items   []*attentionservice.Item
	readAt  time.Time
	nowFunc func() time.Time
}

// New builds the attention service. fx wires the arguments from the provider
// list in registry/provides.go.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(appRepo repository.AppRepo, dockerManager docker.Manager) attentionservice.Service {
	return &service{appRepo: appRepo, dockerManager: dockerManager, nowFunc: time.Now}
}

func (s *service) Items(ctx context.Context, db database.IDB) ([]*attentionservice.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	if s.items != nil && now.Sub(s.readAt) < cacheTTL {
		return s.items, nil
	}
	state, err := s.readCluster(ctx, db, now)
	if err != nil {
		return nil, err
	}
	items := deriveItems(state)
	if items == nil {
		items = []*attentionservice.Item{}
	}
	s.items, s.readAt = items, now
	return items, nil
}

func (s *service) readCluster(ctx context.Context, db database.IDB, now time.Time) (*clusterState, error) {
	apps, _, err := s.appRepo.List(ctx, db, "", nil,
		bunex.SelectColumns("id", "key", "name", "project_id", "project_env_id", "service_id", "status"),
		bunex.SelectWhere("app.service_id IS NOT NULL"),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectRelation("Project", bunex.SelectColumns("id", "key", "name")),
		bunex.SelectRelation("ProjectEnv", bunex.SelectColumns("id", "key")),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	services, err := s.dockerManager.ServiceList(ctx, func(opts *client.ServiceListOptions) {
		opts.Status = true
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	tasks, err := s.dockerManager.TaskList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &clusterState{
		now:      now,
		apps:     apps,
		services: services.Items,
		tasks:    tasks.Items,
		nodes:    nodes.Items,
	}, nil
}
