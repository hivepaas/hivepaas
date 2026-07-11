package clusterserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/events"
)

func (s *service) OnNodeEvent(ctx context.Context, event *events.Message) error {
	return nil
}
