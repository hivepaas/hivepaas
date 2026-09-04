package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/datakeyservice"
)

// InitDataKey installs the key stored secrets are encrypted with, before anything
// can read or write one.
//
// Failing here is deliberate: an app secret that does not open the stored key
// would otherwise surface much later as settings that cannot be decrypted, one
// request at a time.
func InitDataKey(
	lc fx.Lifecycle,
	db *database.DB,
	dataKeyService datakeyservice.Service,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("loading data encryption key...")
			if err := dataKeyService.Load(ctx, db); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
}
