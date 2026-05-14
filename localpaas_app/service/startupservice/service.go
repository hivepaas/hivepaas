package startupservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type Service interface {
	Shutdown()

	LoadLocalPaaSServiceSetting(ctx context.Context) (*entity.Setting, error)
}
