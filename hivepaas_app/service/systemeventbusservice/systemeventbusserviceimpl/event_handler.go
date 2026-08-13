package systemeventbusserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func (s *service) handleEvent(_ context.Context, event *entity.SystemEvent) {
	var err error
	switch event.Type {
	case base.SystemEventHivepaasDomainReload:
		err = s.onHivepaasDomainReload()
	case base.SystemEventPeriodicSettingsReload:
	}
	if err != nil {
		s.logger.Errorf("failed to handle system event %s: %v", event.Type, err)
	}
}
