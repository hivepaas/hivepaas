package startupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type Service interface {
	Shutdown()

	LoadHivePaaSServiceSetting(ctx context.Context) (*entity.Setting, error)
}
