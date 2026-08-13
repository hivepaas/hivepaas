package systemeventbusserviceimpl

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
)

func New(
	client rediscache.Client,
	logger logging.Logger,
) systemeventbusservice.Service {
	return &service{
		client:    client,
		logger:    logger,
		listeners: make(map[base.SystemEventType][]chan *entity.SystemEvent),
	}
}

type service struct {
	client rediscache.Client
	logger logging.Logger

	mu        sync.RWMutex
	listeners map[base.SystemEventType][]chan *entity.SystemEvent

	pubSub     *redis.PubSub
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}
