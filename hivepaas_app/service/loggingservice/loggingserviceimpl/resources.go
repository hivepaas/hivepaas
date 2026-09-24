package loggingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// defaultBackendResources is what a backend gets when nothing says otherwise.
func defaultBackendResources() logging.Resources {
	return logging.Resources{MemoryLimit: logging.DefaultBackendMemoryLimit.Bytes()}
}

// applyBackendResources writes the limits onto the backend's service. The
// service is where they live - the app's own resource screen reads and writes
// the same place - so nothing else holds them.
func (s *service) applyBackendResources(ctx context.Context, app *entity.App, res logging.Resources) error {
	return hperrors.Wrap(s.systemAppService.SetResources(ctx, app, systemappservice.Resources{
		CPULimit: res.CPULimit, MemoryLimit: res.MemoryLimit,
	}))
}

func toLoggingResources(res systemappservice.Resources) logging.Resources {
	return logging.Resources{CPULimit: res.CPULimit, MemoryLimit: res.MemoryLimit}
}
